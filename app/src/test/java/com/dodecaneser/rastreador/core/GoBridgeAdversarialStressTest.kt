package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.coroutines.*
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test
import java.io.IOException

/**
 * Adversarial Stress & Boundary Fuzzing Test Suite for Kotlin Facade (GoBridgeImpl & JvmFallbackBridgeDriver).
 * Milestone 2 Challenger Empirical Verification.
 */
class GoBridgeAdversarialStressTest {

    // ========================================================================
    // 1. Malformed & Corrupted Driver JSON Responses Fuzzing
    // ========================================================================

    @Test
    fun testCorruptedJsonPayloadsFromDriverHandledSafely() = runTest {
        val corruptedPayloads = listOf(
            "",
            "   ",
            "NOT_JSON_AT_ALL",
            "{",
            "}",
            """{"success": true, "data": """,
            """{"success": false""",
            "[]",
            "null",
            "12345",
            """{"success": true, "data": null}""",
            """{"success": true, "data": "not_an_object"}""",
            """{"success": true, "data": {"ip": 12345, "asn": "not_an_int"}}""",
            """{"success": false, "error": null}""",
            """{"success": false, "error": "simulated bridge error message"}""",
            "<xml><error>fatal</error></xml>",
            "\u0000\u0001\u0002"
        )

        for (corrupted in corruptedPayloads) {
            val mockingDriver = object : NativeBridgeDriver {
                override fun isAvailable(): Boolean = true
                override fun queryBgpAsnJson(jsonRequest: String): String = corrupted
                override fun performMultilaterationJson(jsonRequest: String): String = corrupted
                override fun triangulateWiFiJson(jsonRequest: String): String = corrupted
                override fun analyzeIpIdJson(jsonRequest: String): String = corrupted
                override fun analyzeTunnelJson(jsonRequest: String): String = corrupted
                override fun calculateHaversineDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double = 0.0
            }

            val bridge = GoBridgeImpl(driver = mockingDriver, ioDispatcher = Dispatchers.Default)

            // queryBgpAsn must return Result.failure and NEVER throw an uncaught exception
            val bgpRes = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 2)
            assertTrue("Expected failure for corrupted driver output: '$corrupted'", bgpRes.isFailure)
            assertNotNull(bgpRes.exceptionOrNull())

            // performMultilateration
            val multiRes = bridge.performMultilateration("147.96.1.1", timeoutSec = 2)
            assertTrue("Expected failure for corrupted driver output: '$corrupted'", multiRes.isFailure)

            // triangulateWiFi
            val wifiRes = bridge.triangulateWiFi(
                listOf(WiFiBeaconScan("00:11:22:33:44:55", "TestNet", -60, 2412, 1, 20, "[WPA2]", 40.0, -3.0)),
                timeoutSec = 2
            )
            assertTrue("Expected failure for corrupted driver output: '$corrupted'", wifiRes.isFailure)

            // analyzeIpId
            val ipidRes = bridge.analyzeIpId("147.96.1.1", timeoutSec = 2)
            assertTrue("Expected failure for corrupted driver output: '$corrupted'", ipidRes.isFailure)

            // analyzeTunnel
            val tunnelRes = bridge.analyzeTunnel("147.96.1.1", timeoutSec = 2)
            assertTrue("Expected failure for corrupted driver output: '$corrupted'", tunnelRes.isFailure)
        }
    }

    // ========================================================================
    // 2. Driver Native Exception & Crash Containment
    // ========================================================================

    @Test
    fun testDriverExceptionsContainedInResultFailure() = runTest {
        val throwablesToTest = listOf(
            RuntimeException("Simulated native runtime exception"),
            NullPointerException("Simulated null pointer in driver"),
            IllegalStateException("Simulated driver illegal state"),
            IllegalArgumentException("Simulated driver bad argument"),
            IOException("Simulated driver socket failure")
        )

        for (thrown in throwablesToTest) {
            val throwingDriver = object : NativeBridgeDriver {
                override fun isAvailable(): Boolean = true
                override fun queryBgpAsnJson(jsonRequest: String): String = throw thrown
                override fun performMultilaterationJson(jsonRequest: String): String = throw thrown
                override fun triangulateWiFiJson(jsonRequest: String): String = throw thrown
                override fun analyzeIpIdJson(jsonRequest: String): String = throw thrown
                override fun analyzeTunnelJson(jsonRequest: String): String = throw thrown
                override fun calculateHaversineDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double = throw thrown
            }

            val bridge = GoBridgeImpl(driver = throwingDriver, ioDispatcher = Dispatchers.Default)

            val bgpRes = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 2)
            assertTrue("Expected Result.failure when driver throws ${thrown::class.simpleName}", bgpRes.isFailure)
            assertEquals(thrown::class, bgpRes.exceptionOrNull()!!::class)
            assertEquals(thrown.message, bgpRes.exceptionOrNull()?.message)
        }
    }

    // ========================================================================
    // 3. Out-of-Bounds Geographic Inputs (lat > 90, lon > 180, NaN, Inf)
    // ========================================================================

    @Test
    fun testHaversineGeographicBoundariesAndSingularities() {
        val driver = JvmFallbackBridgeDriver()
        val bridge = GoBridgeImpl(driver = driver)

        // Equator half-circumference (~20015 km)
        val equatorHalf = bridge.calculateHaversineDistance(0.0, 0.0, 0.0, 180.0)
        assertEquals(20015.08, equatorHalf, 50.0)

        // North pole to South pole (~20015 km)
        val poleToPole = bridge.calculateHaversineDistance(90.0, 0.0, -90.0, 0.0)
        assertEquals(20015.08, poleToPole, 50.0)

        // Zero distance (identical points)
        val zero = bridge.calculateHaversineDistance(40.4168, -3.7038, 40.4168, -3.7038)
        assertEquals(0.0, zero, 0.0001)

        // Beyond poles (> 90 lat) - must not crash
        val beyondPoles = bridge.calculateHaversineDistance(95.0, 0.0, 100.0, 0.0)
        assertFalse("Distance must not be NaN for finite input", beyondPoles.isNaN())

        // Beyond dateline (> 180 lon) - must not crash
        val beyondDateline = bridge.calculateHaversineDistance(0.0, 200.0, 0.0, 250.0)
        assertFalse("Distance must not be NaN for finite input", beyondDateline.isNaN())

        // IEEE 754 NaN handling
        val nanDist = bridge.calculateHaversineDistance(Double.NaN, 0.0, 0.0, 0.0)
        assertTrue("Distance with NaN input must be NaN", nanDist.isNaN())

        // IEEE 754 Infinity handling
        val infDist = bridge.calculateHaversineDistance(Double.POSITIVE_INFINITY, 0.0, 0.0, 0.0)
        assertTrue("Distance with Inf input must be NaN", infDist.isNaN())
    }

    // ========================================================================
    // 4. Empty, Invalid, and Private IPv4/IPv6 Address Strings
    // ========================================================================

    @Test
    fun testEmptyAndInvalidIpAddresses() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)

        // Empty string
        val emptyRes = bridge.queryBgpAsn("", timeoutSec = 2)
        assertTrue("Empty IP must fail", emptyRes.isFailure)
        assertTrue(emptyRes.exceptionOrNull()?.message?.contains("ip parameter is required") == true)

        // Whitespace string
        val blankRes = bridge.queryBgpAsn("   ", timeoutSec = 2)
        assertTrue("Blank IP must fail", blankRes.isFailure)

        // Alphanumeric invalid IP
        val invalidRes = bridge.queryBgpAsn("invalid-ip", timeoutSec = 2)
        assertTrue("Invalid IP must fail", invalidRes.isFailure)
        assertTrue(invalidRes.exceptionOrNull()?.message?.contains("invalid IP address") == true)

        // Octets > 255
        val outOfBoundsIp = bridge.queryBgpAsn("999.999.999.999", timeoutSec = 2)
        assertTrue("Octets > 255 must fail", outOfBoundsIp.isFailure)

        // Incomplete IP
        val incompleteIp = bridge.queryBgpAsn("1.2.3", timeoutSec = 2)
        assertTrue("Incomplete IP must fail", incompleteIp.isFailure)

        // Private RFC1918 IPs must succeed with synthesized/resolved metadata
        val privateA = bridge.queryBgpAsn("10.0.0.1", timeoutSec = 2)
        assertTrue("10.0.0.1 must succeed", privateA.isSuccess)
        assertEquals("LOCAL", privateA.getOrNull()?.countryCode)

        val privateB = bridge.queryBgpAsn("172.16.0.1", timeoutSec = 2)
        assertTrue("172.16.0.1 must succeed", privateB.isSuccess)

        val privateC = bridge.queryBgpAsn("192.168.1.1", timeoutSec = 2)
        assertTrue("192.168.1.1 must succeed", privateC.isSuccess)

        val loopback = bridge.queryBgpAsn("127.0.0.1", timeoutSec = 2)
        assertTrue("127.0.0.1 must succeed", loopback.isSuccess)
    }

    // ========================================================================
    // 5. WiFi Beacons Boundary Cases (zero, single, negative/zero RSSI, identical coords)
    // ========================================================================

    @Test
    fun testWiFiZeroBeacons() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val res = bridge.triangulateWiFi(emptyList(), timeoutSec = 2)
        assertTrue("Zero beacons must succeed with 0 resolved count", res.isSuccess)
        val data = res.getOrNull()
        assertNotNull(data)
        assertEquals(0, data?.resolvedCount)
        assertEquals(0, data?.totalBeacons)
    }

    @Test
    fun testWiFiSingleBeacon() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val beacon = WiFiBeaconScan(
            bssid = "00:11:22:33:44:55",
            ssid = "SoloAP",
            rssi = -55,
            frequency = 2412,
            channel = 1,
            channelWidth = 20,
            capabilities = "[WPA2]",
            lat = 40.4168,
            lon = -3.7038
        )
        val res = bridge.triangulateWiFi(listOf(beacon), timeoutSec = 2)
        assertTrue(res.isSuccess)
        val data = res.getOrNull()
        assertNotNull(data)
        assertEquals(1, data?.resolvedCount)
        assertEquals(40.4168, data?.estimatedPoint?.lat ?: 0.0, 0.001)
        assertEquals(-3.7038, data?.estimatedPoint?.lon ?: 0.0, 0.001)
    }

    @Test
    fun testWiFiNegativeAndZeroRSSI() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val beacons = listOf(
            WiFiBeaconScan("00:11:22:33:44:01", "AP1", rssi = 0, lat = 40.4168, lon = -3.7038),
            WiFiBeaconScan("00:11:22:33:44:02", "AP2", rssi = -120, lat = 40.4170, lon = -3.7040),
            WiFiBeaconScan("00:11:22:33:44:03", "AP3", rssi = -30, lat = 40.4169, lon = -3.7039)
        )
        val res = bridge.triangulateWiFi(beacons, timeoutSec = 2)
        assertTrue(res.isSuccess)
        val data = res.getOrNull()
        assertNotNull(data)
        assertFalse(data?.estimatedPoint?.lat?.isNaN() == true)
        assertFalse(data?.estimatedPoint?.lon?.isNaN() == true)
        assertTrue((data?.precisionM ?: 0.0) > 0.0)
    }

    @Test
    fun testWiFiIdenticalCoordinates() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)
        val beacons = listOf(
            WiFiBeaconScan("00:11:22:33:44:01", "AP1", rssi = -50, lat = 40.4168, lon = -3.7038),
            WiFiBeaconScan("00:11:22:33:44:02", "AP2", rssi = -60, lat = 40.4168, lon = -3.7038),
            WiFiBeaconScan("00:11:22:33:44:03", "AP3", rssi = -70, lat = 40.4168, lon = -3.7038)
        )
        val res = bridge.triangulateWiFi(beacons, timeoutSec = 2)
        assertTrue(res.isSuccess)
        val data = res.getOrNull()
        assertNotNull(data)
        assertEquals(40.4168, data?.estimatedPoint?.lat ?: 0.0, 0.0001)
        assertEquals(-3.7038, data?.estimatedPoint?.lon ?: 0.0, 0.0001)
        // Zero dispersion should clamp precision to minimum
        assertEquals(8.0, data?.precisionM ?: 0.0, 0.001)
    }

    // ========================================================================
    // 6. Extreme Timeouts (0s, Negative, Very Large)
    // ========================================================================

    @Test
    fun testExtremeTimeoutsHandledProperly() = runTest {
        val bridge = GoBridgeImpl(driver = JvmFallbackBridgeDriver(), ioDispatcher = Dispatchers.Default)

        // 0s timeout
        val zeroTimeout = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 0)
        assertTrue(zeroTimeout.isFailure)
        assertTrue(zeroTimeout.exceptionOrNull() is IllegalArgumentException)

        // Negative timeouts
        val neg1 = bridge.queryBgpAsn("147.96.1.1", timeoutSec = -1)
        assertTrue(neg1.isFailure)
        assertTrue(neg1.exceptionOrNull() is IllegalArgumentException)

        val neg100 = bridge.queryBgpAsn("147.96.1.1", timeoutSec = -100)
        assertTrue(neg100.isFailure)
        assertTrue(neg100.exceptionOrNull() is IllegalArgumentException)

        // Very large timeout (100,000s) - must not overflow and execute normally
        val largeTimeout = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 100000)
        assertTrue(largeTimeout.isSuccess)

        // Int.MAX_VALUE (2147483647) - must not overflow Long multiplication
        val maxIntTimeout = bridge.queryBgpAsn("147.96.1.1", timeoutSec = Int.MAX_VALUE)
        assertTrue(maxIntTimeout.isSuccess)
    }
}
