package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.coroutines.*
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test
import java.io.IOException
import java.util.concurrent.atomic.AtomicInteger

class GoBridgeImplThreadingTest {

    @Test
    fun testSuccessfulExecutionReturnsSuccess() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 5)

        assertTrue("Expected Result.isSuccess == true", result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals("147.96.1.1", data?.ip)
        assertEquals(766, data?.asn)
        assertEquals("Universidad Complutense de Madrid", data?.asOrg)
    }

    @Test
    fun testZeroOrNegativeTimeoutImmediateFailure() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val resultZero = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 0)
        assertTrue("Zero timeout must fail", resultZero.isFailure)
        assertTrue(resultZero.exceptionOrNull() is IllegalArgumentException)

        val resultNegative = bridge.queryBgpAsn("147.96.1.1", timeoutSec = -2)
        assertTrue("Negative timeout must fail", resultNegative.isFailure)
        assertTrue(resultNegative.exceptionOrNull() is IllegalArgumentException)
    }

    @Test
    fun testTimeoutExpiresReturnsResultFailure() = runBlocking(Dispatchers.Default) {
        val slowDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                Thread.sleep(1500)
                return JvmFallbackBridgeDriver().queryBgpAsnJson(jsonRequest)
            }
        }

        val bridge = GoBridgeImpl(driver = slowDriver, ioDispatcher = Dispatchers.IO)
        val result = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 1)

        assertTrue("Expected failure on timeout expiry", result.isFailure)
        assertTrue(
            "Expected TimeoutCancellationException, got: ${result.exceptionOrNull()}",
            result.exceptionOrNull() is TimeoutCancellationException
        )
    }

    @Test
    fun testCancellationExceptionPropagated() = runTest {
        val slowDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                Thread.sleep(200)
                return JvmFallbackBridgeDriver().queryBgpAsnJson(jsonRequest)
            }
        }
        val bridge = GoBridgeImpl(driver = slowDriver, ioDispatcher = Dispatchers.IO)

        var cancellationCaught = false
        val job = launch(Dispatchers.Default) {
            try {
                bridge.queryBgpAsn("147.96.1.1", timeoutSec = 5)
            } catch (e: CancellationException) {
                cancellationCaught = true
                throw e
            }
        }

        delay(50)
        job.cancelAndJoin()

        assertTrue(
            "CancellationException must be propagated cooperatively and not swallowed into Result.failure",
            cancellationCaught
        )
    }

    @Test
    fun testBridgeDriverExceptionWrappedInResultFailure() = runTest {
        val failingDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                throw IOException("Native socket reset")
            }
        }

        val bridge = GoBridgeImpl(driver = failingDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 5)

        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IOException)
        assertEquals("Native socket reset", result.exceptionOrNull()?.message)
    }

    @Test
    fun testBridgeDriverErrorEnvelopeWrappedInResultFailure() = runTest {
        val errorEnvelopeDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                return """{"success": false, "error": "target unreachable"}"""
            }
        }

        val bridge = GoBridgeImpl(driver = errorEnvelopeDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 5)

        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull()?.message?.contains("target unreachable") == true)
    }

    @Test
    fun testConcurrentQueriesThreadSafety() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val counter = AtomicInteger(0)

        val jobs = (1..20).map { i ->
            async(Dispatchers.Default) {
                val res = bridge.queryBgpAsn("147.96.1.$i", timeoutSec = 5)
                if (res.isSuccess) {
                    counter.incrementAndGet()
                }
            }
        }

        jobs.awaitAll()
        assertEquals(20, counter.get())
    }

    @Test
    fun testAllSixOperationsViaBridge() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)

        // 1. queryBgpAsn
        val bgp = bridge.queryBgpAsn("8.8.8.8", timeoutSec = 5)
        assertTrue(bgp.isSuccess)
        assertEquals("Google LLC", bgp.getOrNull()?.asOrg)

        // 2. performMultilateration
        val multilat = bridge.performMultilateration("147.96.1.1", timeoutSec = 5)
        assertTrue("Multilat failed: ${multilat.exceptionOrNull()}", multilat.isSuccess)
        assertTrue((multilat.getOrNull()?.confidenceKm ?: 0.0) > 0.0)

        // 3. triangulateWiFi
        val wifi = bridge.triangulateWiFi(
            listOf(
                WiFiBeaconScan("00:11:22:33:44:55", "TestNet", -60, 2412, 1, 20, "[WPA2]", 40.4168, -3.7038)
            ),
            timeoutSec = 5
        )
        assertTrue("Wifi failed: ${wifi.exceptionOrNull()}", wifi.isSuccess)
        assertEquals(1, wifi.getOrNull()?.resolvedCount)

        // 4. analyzeIpId
        val ipid = bridge.analyzeIpId("147.96.1.1", timeoutSec = 5)
        assertTrue("IpId failed: ${ipid.exceptionOrNull()}", ipid.isSuccess)
        assertEquals("GLOBAL_INCREMENTAL", ipid.getOrNull()?.generationType)

        // 5. analyzeTunnel
        val tunnel = bridge.analyzeTunnel("10.0.0.1", l7Url = "https://example.com", timeoutSec = 5)
        assertTrue("Tunnel failed: ${tunnel.exceptionOrNull()}", tunnel.isSuccess)
        assertTrue(tunnel.getOrNull()?.isTunnelDetected == true)

        // 6. calculateHaversineDistance
        val dist = bridge.calculateHaversineDistance(40.4168, -3.7038, 41.3851, 2.1734)
        assertTrue("Distance between Madrid and Barcelona should be ~505 km", dist in 500.0..515.0)
    }
}
