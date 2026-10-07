package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.coroutines.*
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test
import java.io.IOException
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger
import kotlin.random.Random

/**
 * Empirical Challenger Test Harness:
 * Stress-testing GoBridgeImpl for:
 * 1. 50+ concurrent coroutines making simultaneous multi-operation calls.
 * 2. Rapid cancellation of coroutine jobs mid-execution (asserting cooperative CancellationException).
 * 3. Timeout deadline enforcement under simulated network latency and concurrent timeout races.
 * 4. High-throughput memory allocation and GC stability across repeated bridge calls.
 * 5. Driver failure containment and large-payload boundary stress.
 */
class GoBridgeConcurrencyStressTest {

    // ----------------------------------------------------------------------------------
    // 1. High-Concurrency Stress Test (100 Concurrent Coroutines Across Mixed Operations)
    // ----------------------------------------------------------------------------------
    @Test
    fun test100ConcurrentMixedOperationsStability() = runBlocking {
        val customDispatcher = Executors.newFixedThreadPool(16).asCoroutineDispatcher()
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = customDispatcher)

        val coroutineCount = 100
        val successCount = AtomicInteger(0)
        val failureMap = ConcurrentHashMap<Int, Throwable>()
        val resultsMap = ConcurrentHashMap<Int, Any>()

        coroutineScope {
            val jobs = (1..coroutineCount).map { id ->
                launch(customDispatcher) {
                    try {
                        when (id % 5) {
                            0 -> {
                                val res = bridge.queryBgpAsn("147.96.1.${(id % 250) + 1}", timeoutSec = 5)
                                if (res.isSuccess) {
                                    successCount.incrementAndGet()
                                    resultsMap[id] = res.getOrThrow()
                                } else {
                                    failureMap[id] = res.exceptionOrNull() ?: Exception("Unknown error")
                                }
                            }
                            1 -> {
                                val res = bridge.performMultilateration("147.96.1.${(id % 250) + 1}", timeoutSec = 5)
                                if (res.isSuccess) {
                                    successCount.incrementAndGet()
                                    resultsMap[id] = res.getOrThrow()
                                } else {
                                    failureMap[id] = res.exceptionOrNull() ?: Exception("Unknown error")
                                }
                            }
                            2 -> {
                                val beacons = listOf(
                                    WiFiBeaconScan("00:11:22:33:44:${(id % 90) + 10}", "AP-$id", -50 - (id % 30), 2412, 1, 20, "[WPA2]")
                                )
                                val res = bridge.triangulateWiFi(beacons, timeoutSec = 5)
                                if (res.isSuccess) {
                                    successCount.incrementAndGet()
                                    resultsMap[id] = res.getOrThrow()
                                } else {
                                    failureMap[id] = res.exceptionOrNull() ?: Exception("Unknown error")
                                }
                            }
                            3 -> {
                                val res = bridge.analyzeIpId("147.96.1.${(id % 250) + 1}", timeoutSec = 5)
                                if (res.isSuccess) {
                                    successCount.incrementAndGet()
                                    resultsMap[id] = res.getOrThrow()
                                } else {
                                    failureMap[id] = res.exceptionOrNull() ?: Exception("Unknown error")
                                }
                            }
                            4 -> {
                                val res = bridge.analyzeTunnel("10.0.0.${(id % 250) + 1}", l7Url = "https://server-$id.local", timeoutSec = 5)
                                if (res.isSuccess) {
                                    successCount.incrementAndGet()
                                    resultsMap[id] = res.getOrThrow()
                                } else {
                                    failureMap[id] = res.exceptionOrNull() ?: Exception("Unknown error")
                                }
                            }
                        }
                    } catch (t: Throwable) {
                        failureMap[id] = t
                    }
                }
            }
            jobs.joinAll()
        }

        customDispatcher.close()

        assertTrue(
            "Expected 0 failures in 100 concurrent coroutines, but got ${failureMap.size}: ${failureMap.entries.take(5)}",
            failureMap.isEmpty()
        )
        assertEquals("All 100 concurrent coroutines must succeed", coroutineCount, successCount.get())
        assertEquals("Result map must contain 100 entries", coroutineCount, resultsMap.size)
    }

    // ----------------------------------------------------------------------------------
    // 2. Rapid Coroutine Cancellation Stress (Mid-Execution Cooperative Cancellation)
    // ----------------------------------------------------------------------------------
    @Test
    fun testRapidCancellationMidExecutionCooperativePropagation() = runBlocking {
        // Driver simulating latency of 300ms per call
        val delayedDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                Thread.sleep(300)
                return JvmFallbackBridgeDriver().queryBgpAsnJson(jsonRequest)
            }
        }

        val bridge = GoBridgeImpl(driver = delayedDriver, ioDispatcher = Dispatchers.IO)
        val cancelAttempts = 20
        val cancellationPropagatedCount = AtomicInteger(0)
        val swallowedAsResultFailureCount = AtomicInteger(0)
        val completedCount = AtomicInteger(0)

        coroutineScope {
            val jobs = (1..cancelAttempts).map { i ->
                val delayBeforeCancel = (i * 5L) // 5ms, 10ms, ... 100ms
                launch(Dispatchers.Default) {
                    val childJob = launch(Dispatchers.IO) {
                        try {
                            val res = bridge.queryBgpAsn("147.96.1.$i", timeoutSec = 5)
                            if (res.isFailure && res.exceptionOrNull() is CancellationException) {
                                swallowedAsResultFailureCount.incrementAndGet()
                            } else {
                                completedCount.incrementAndGet()
                            }
                        } catch (e: CancellationException) {
                            cancellationPropagatedCount.incrementAndGet()
                            throw e
                        }
                    }

                    delay(delayBeforeCancel)
                    childJob.cancel()
                    childJob.join()
                }
            }
            jobs.joinAll()
        }

        assertEquals(
            "CancellationException must NEVER be swallowed into Result.failure",
            0,
            swallowedAsResultFailureCount.get()
        )
        assertTrue(
            "At least one cancellation should have been caught during active execution (got $cancellationPropagatedCount)",
            cancellationPropagatedCount.get() > 0
        )
    }

    @Test
    fun testImmediatePreExecutionCancellation() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.IO)
        val wasInvoked = AtomicBoolean(false)

        val job = launch(Dispatchers.IO) {
            // Already canceled before entering bridge
            cancel()
            val result = bridge.queryBgpAsn("8.8.8.8", timeoutSec = 5)
            wasInvoked.set(true)
        }

        job.join()
        assertTrue("Job must be cancelled", job.isCancelled)
        assertFalse("Bridge operation must not complete when pre-cancelled", wasInvoked.get())
    }

    // ----------------------------------------------------------------------------------
    // 3. Timeout Deadline Enforcement & Concurrent Races
    // ----------------------------------------------------------------------------------
    @Test
    fun testStrictTimeoutEnforcementUnderSimulatedLatency() = runBlocking(Dispatchers.Default) {
        val slowDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                Thread.sleep(1200) // Exceeds 1s timeout
                return JvmFallbackBridgeDriver().queryBgpAsnJson(jsonRequest)
            }
        }

        val bridge = GoBridgeImpl(driver = slowDriver, ioDispatcher = Dispatchers.IO)
        val startTime = System.currentTimeMillis()
        val result = bridge.queryBgpAsn("8.8.8.8", timeoutSec = 1)
        val elapsed = System.currentTimeMillis() - startTime

        assertTrue("Call exceeding timeout must result in failure", result.isFailure)
        assertTrue(
            "Failure cause must be TimeoutCancellationException, but was: ${result.exceptionOrNull()}",
            result.exceptionOrNull() is TimeoutCancellationException
        )
        // Elapsed time should be bounded close to timeout (around 1000ms - 1500ms)
        assertTrue("Execution should be timed out in reasonable duration, elapsed: ${elapsed}ms", elapsed in 900..1800)
    }

    @Test
    fun testConcurrentTimeoutRaceConditionIsolation() = runBlocking {
        // Driver where even IDs take 1500ms (timeout) and odd IDs take 50ms (fast success)
        val asymmetricDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String {
                if (jsonRequest.contains("147.96.1.even")) {
                    Thread.sleep(1500)
                } else {
                    Thread.sleep(50)
                }
                return JvmFallbackBridgeDriver().queryBgpAsnJson(jsonRequest.replace("even", "1").replace("odd", "2"))
            }
        }

        val bridge = GoBridgeImpl(driver = asymmetricDriver, ioDispatcher = Dispatchers.IO)
        val fastSuccesses = AtomicInteger(0)
        val slowTimeouts = AtomicInteger(0)
        val unexpectedErrors = ConcurrentHashMap<Int, Throwable>()

        coroutineScope {
            val jobs = (1..40).map { i ->
                launch(Dispatchers.Default) {
                    val isEven = (i % 2 == 0)
                    val ip = if (isEven) "147.96.1.even" else "147.96.1.odd"
                    val res = bridge.queryBgpAsn(ip, timeoutSec = 1)

                    if (isEven) {
                        if (res.isFailure && res.exceptionOrNull() is TimeoutCancellationException) {
                            slowTimeouts.incrementAndGet()
                        } else {
                            unexpectedErrors[i] = res.exceptionOrNull() ?: Exception("Expected timeout failure for even, got: $res")
                        }
                    } else {
                        if (res.isSuccess) {
                            fastSuccesses.incrementAndGet()
                        } else {
                            unexpectedErrors[i] = res.exceptionOrNull() ?: Exception("Expected success for odd, got: $res")
                        }
                    }
                }
            }
            jobs.joinAll()
        }

        assertTrue("No cross-contamination errors between concurrent timeout races: $unexpectedErrors", unexpectedErrors.isEmpty())
        assertEquals("All 20 fast coroutines must succeed", 20, fastSuccesses.get())
        assertEquals("All 20 slow coroutines must time out independently", 20, slowTimeouts.get())
    }

    // ----------------------------------------------------------------------------------
    // 4. Memory Allocation & GC Stability Under High-Throughput Repeated Calls
    // ----------------------------------------------------------------------------------
    @Test
    fun testMemoryAllocationAndGCStabilityAcross1000Calls() = runBlocking {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)

        // Warm up JVM and JIT
        repeat(50) {
            bridge.calculateHaversineDistance(40.4168, -3.7038, 41.3851, 2.1734)
            bridge.queryBgpAsn("127.0.0.1", timeoutSec = 5)
        }

        System.gc()
        Thread.sleep(100)
        val runtime = Runtime.getRuntime()
        val memBeforeBytes = runtime.totalMemory() - runtime.freeMemory()

        val iterationCount = 1000
        val counter = AtomicInteger(0)

        // Run 1000 rapid bridge calls in batches of 50 concurrent coroutines
        val batchSize = 50
        for (batch in 0 until (iterationCount / batchSize)) {
            coroutineScope {
                val jobs = (0 until batchSize).map { idx ->
                    val callId = batch * batchSize + idx
                    launch(Dispatchers.Default) {
                        val res = bridge.queryBgpAsn("147.96.1.${(callId % 200) + 1}", timeoutSec = 5)
                        if (res.isSuccess) {
                            counter.incrementAndGet()
                        }
                    }
                }
                jobs.joinAll()
            }
        }

        assertEquals("All 1000 requests must complete successfully", iterationCount, counter.get())

        System.gc()
        Thread.sleep(100)
        val memAfterBytes = runtime.totalMemory() - runtime.freeMemory()
        val deltaMb = (memAfterBytes - memBeforeBytes) / (1024.0 * 1024.0)

        // Verifying that after GC, heap retention increase is reasonable (< 35 MB) for 1000 JSON serialization passes
        assertTrue(
            "Memory growth after GC must remain bounded (delta: ${deltaMb}MB)",
            deltaMb < 35.0
        )
    }

    // ----------------------------------------------------------------------------------
    // 5. Adversarial Driver Exceptions & Large Payload Boundary Tests
    // ----------------------------------------------------------------------------------
    @Test
    fun testCorruptedJsonDriverResponseHandledGracefully() = runTest {
        val malformedDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String =
                "{{{MALFORMED_JSON_STRING_FROM_NATIVE_CRASH"
        }

        val bridge = GoBridgeImpl(driver = malformedDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 5)

        assertTrue("Malformed JSON must result in Result.failure", result.isFailure)
        assertNotNull("Exception must not be null", result.exceptionOrNull())
    }

    @Test
    fun testDriverNativeRuntimeExceptionContainment() = runTest {
        val crashingDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun performMultilaterationJson(jsonRequest: String): String {
                throw RuntimeException("Fatal JNI memory access violation simulated")
            }
        }

        val bridge = GoBridgeImpl(driver = crashingDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.performMultilateration("147.96.1.1", timeoutSec = 5)

        assertTrue("Native driver runtime exception must be trapped into Result.failure", result.isFailure)
        assertEquals("Fatal JNI memory access violation simulated", result.exceptionOrNull()?.message)
    }

    @Test
    fun testLargeBeaconArrayPayloadStress() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)

        // Generate 500 beacons
        val beacons = (1..500).map { i ->
            val bssid = String.format("%02x:%02x:%02x:%02x:%02x:%02x", i % 256, (i / 256) % 256, 1, 2, 3, 4)
            WiFiBeaconScan(
                bssid = bssid,
                ssid = "Large-Scale-WiFi-$i",
                rssi = -30 - (i % 60),
                frequency = 2412 + (i % 10) * 5,
                channel = (i % 11) + 1,
                channelWidth = 20,
                capabilities = "[WPA2-PSK-CCMP][ESS]",
                lat = 40.4168 + (i % 100) * 0.0001,
                lon = -3.7038 + (i % 100) * 0.0001,
                timestamp = System.currentTimeMillis()
            )
        }

        val result = bridge.triangulateWiFi(beacons, timeoutSec = 10)
        assertTrue("Large beacon dataset must succeed: ${result.exceptionOrNull()}", result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals(500, data?.totalBeacons)
        assertEquals(500, data?.resolvedCount)
        assertTrue("Estimated latitude should be in Madrid area", data!!.estimatedPoint.lat in 40.4..40.5)
        assertTrue("Estimated longitude should be in Madrid area", data.estimatedPoint.lon in -3.8..-3.6)
    }
}
