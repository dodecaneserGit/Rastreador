package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Test

class JvmFallbackBridgeDriverTest {

    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun testJniDriverGracefulWhenNativeMissing() {
        val jni = JniNativeBridgeDriver()
        // In desktop JVM unit tests, librastreador is not linked into the host JVM
        assertFalse("JniNativeBridgeDriver must report false when library is missing", jni.isAvailable())
    }

    @Test
    fun testDefaultDriverResolvesFallbackOnDesktop() {
        val driver = GoBridgeImpl.resolveDefaultDriver()
        assertNotNull(driver)
        assertTrue("Resolved driver must report available", driver.isAvailable())
        assertTrue("Desktop environment must resolve to JvmFallbackBridgeDriver", driver is JvmFallbackBridgeDriver)
    }

    @Test
    fun testJvmFallbackQueryBgpAsnSuccess() {
        val driver = JvmFallbackBridgeDriver()
        val rawJson = driver.queryBgpAsnJson("""{"ip": "147.96.1.1", "timeout_sec": 5}""")

        val env = json.decodeFromString<BridgeEnvelope<BgpAsnResult>>(rawJson)
        assertTrue(env.success)
        assertNotNull(env.data)
        assertEquals("147.96.1.1", env.data?.ip)
        assertEquals(766, env.data?.asn)
        assertEquals("Universidad Complutense de Madrid", env.data?.asOrg)
        assertEquals("Madrid", env.data?.city)
        assertEquals("ES", env.data?.country)
    }

    @Test
    fun testJvmFallbackQueryBgpAsnInvalidIP() {
        val driver = JvmFallbackBridgeDriver()
        val rawJson = driver.queryBgpAsnJson("""{"ip": "999.999.999.999"}""")

        val env = json.decodeFromString<BridgeEnvelope<BgpAsnResult>>(rawJson)
        assertFalse(env.success)
        assertNotNull(env.error)
        assertTrue(env.error?.contains("invalid IP") == true)
    }

    @Test
    fun testJvmFallbackMultilateration() {
        val driver = JvmFallbackBridgeDriver()
        val rawJson = driver.performMultilaterationJson("""{"ip": "147.96.1.1", "timeout_sec": 5}""")

        val env = json.decodeFromString<BridgeEnvelope<MultilaterationResult>>(rawJson)
        assertTrue(env.success)
        assertNotNull(env.data)
        assertTrue((env.data?.confidenceKm ?: 0.0) > 0.0)
        assertTrue((env.data?.usedLandmarks?.size ?: 0) >= 3)
    }

    @Test
    fun testJvmFallbackWiFiTriangulationWithBeacons() {
        val driver = JvmFallbackBridgeDriver()
        val req = """
            {
                "beacons": [
                    {"bssid": "00:11:22:33:44:01", "rssi": -40, "latitude": 40.4168, "longitude": -3.7038},
                    {"bssid": "00:11:22:33:44:02", "rssi": -90, "latitude": 40.4180, "longitude": -3.7050}
                ]
            }
        """.trimIndent()

        val rawJson = driver.triangulateWiFiJson(req)
        val env = json.decodeFromString<BridgeEnvelope<TriangulationResult>>(rawJson)

        assertTrue(env.success)
        assertNotNull(env.data)
        assertEquals(2, env.data?.resolvedCount)
        assertTrue((env.data?.confidenceKm ?: 0.0) > 0.0)

        // The beacon with -40 dBm has 10^(-40/20) = 0.01 weight; -90 dBm has 10^(-90/20) = 0.00003 weight
        // Therefore estimated point should be very close to beacon 1 (40.4168, -3.7038)
        val estLat = env.data?.estimatedPoint?.lat ?: 0.0
        val estLon = env.data?.estimatedPoint?.lon ?: 0.0
        val distToBeacon1 = driver.calculateHaversineDistance(estLat, estLon, 40.4168, -3.7038)
        assertTrue("Estimated point should be within 0.05 km of stronger beacon", distToBeacon1 < 0.05)
    }

    @Test
    fun testJvmFallbackIpIdAnalysis() {
        val driver = JvmFallbackBridgeDriver()
        val rawJson = driver.analyzeIpIdJson("""{"ip": "147.96.1.1", "samples": 6}""")

        val env = json.decodeFromString<BridgeEnvelope<IpIdResult>>(rawJson)
        assertTrue(env.success)
        assertNotNull(env.data)
        assertEquals("GLOBAL_INCREMENTAL", env.data?.generationType)
        assertTrue((env.data?.velocityPacketsPerSec ?: 0.0) > 0.0)
        assertTrue((env.data?.linearityScore ?: 0.0) > 0.90)
        assertTrue((env.data?.correlationFingerprint?.length ?: 0) >= 8)
    }

    @Test
    fun testJvmFallbackTunnelAnalysis() {
        val driver = JvmFallbackBridgeDriver()
        val rawJson = driver.analyzeTunnelJson("""{"ip": "10.0.0.1", "l7_url": "https://target.local"}""")

        val env = json.decodeFromString<BridgeEnvelope<TunnelResult>>(rawJson)
        assertTrue(env.success)
        assertNotNull(env.data)
        assertTrue(env.data?.isTunnelDetected == true)
        assertEquals("WIREGUARD", env.data?.tunnelType)
        assertTrue(env.data?.mssClampingDetected == true)
        assertEquals("HIGH", env.data?.confidence)
    }
}
