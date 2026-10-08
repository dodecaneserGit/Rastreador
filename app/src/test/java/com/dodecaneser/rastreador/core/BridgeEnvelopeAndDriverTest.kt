package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test
import java.io.IOException

/**
 * Unit Test Suite verifying Envelope Unwrapping Bug Fix, Raw Object Fallback,
 * and Driver Resolution / Graceful Degradation Architecture.
 */
class BridgeEnvelopeAndDriverTest {

    // ========================================================================
    // 1. Envelope Unwrapping & Raw JSON Fallback Tests (Bug Fix Verification)
    // ========================================================================

    @Test
    fun testDirectRawObjectUnwrappingBgpAsn() = runTest {
        // Driver returns raw BgpAsnResult JSON without BridgeEnvelope wrapper
        val rawJson = """
            {
                "ip": "147.96.1.1",
                "asn": 766,
                "as_org": "Universidad Complutense de Madrid",
                "country": "ES",
                "city": "Madrid",
                "latitude": 40.4489,
                "longitude": -3.7297,
                "confidence": "HIGH"
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String = rawJson
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("147.96.1.1", timeoutSec = 5)

        assertTrue("Direct raw JSON must parse successfully into Result.success", result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals("147.96.1.1", data?.ip)
        assertEquals(766, data?.asn)
        assertEquals("Universidad Complutense de Madrid", data?.asOrg)
        assertEquals("Madrid", data?.city)
    }

    @Test
    fun testDirectRawObjectUnwrappingMultilateration() = runTest {
        val rawJson = """
            {
                "estimated_point": {"lat": 40.4168, "lon": -3.7038},
                "confidence_km": 12.5,
                "used_landmarks": [
                    {
                        "id": "POP-MAD",
                        "name": "Madrid Hub",
                        "city": "Madrid",
                        "country": "ES",
                        "location": {"lat": 40.4168, "lon": -3.7038},
                        "min_rtt_ms": 2.1,
                        "max_radius_km": 210.0,
                        "samples": 3,
                        "type": "hop_landmark"
                    }
                ],
                "polygon_bounds": []
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun performMultilaterationJson(jsonRequest: String): String = rawJson
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.performMultilateration("147.96.1.1", timeoutSec = 5)

        assertTrue(result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals(40.4168, data?.estimatedPoint?.lat ?: 0.0, 0.001)
        assertEquals(12.5, data?.confidenceKm ?: 0.0, 0.001)
        assertEquals(1, data?.usedLandmarks?.size)
    }

    @Test
    fun testDirectRawObjectUnwrappingWiFi() = runTest {
        val rawJson = """
            {
                "estimated_point": {"lat": 40.4168, "lon": -3.7038},
                "confidence_km": 0.015,
                "precision_meters": 15.0,
                "resolved_count": 1,
                "total_beacons": 1,
                "street_address": "Gran Vía, Madrid",
                "networks": []
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun triangulateWiFiJson(jsonRequest: String): String = rawJson
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.triangulateWiFi(emptyList(), timeoutSec = 5)

        assertTrue(result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals(15.0, data?.precisionM ?: 0.0, 0.001)
        assertEquals("Gran Vía, Madrid", data?.streetAddress)
    }

    @Test
    fun testDirectRawObjectWithUnknownKeysForwardCompatibility() = runTest {
        val rawJsonWithUnknownKeys = """
            {
                "ip": "8.8.8.8",
                "asn": 15169,
                "as_org": "Google LLC",
                "country": "US",
                "experimental_kernel_probe_data": "0xDEADBEEF",
                "bgp_as_path_depth": 3
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String = rawJsonWithUnknownKeys
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("8.8.8.8", timeoutSec = 5)

        assertTrue(result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals("8.8.8.8", data?.ip)
        assertEquals(15169, data?.asn)
    }

    @Test
    fun testStandardEnvelopeSuccessUnwrapping() = runTest {
        val envelopeJson = """
            {
                "success": true,
                "data": {
                    "ip": "1.1.1.1",
                    "asn": 13335,
                    "as_org": "Cloudflare, Inc.",
                    "country": "US"
                },
                "error": null
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String = envelopeJson
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("1.1.1.1", timeoutSec = 5)

        assertTrue(result.isSuccess)
        val data = result.getOrNull()
        assertNotNull(data)
        assertEquals("1.1.1.1", data?.ip)
        assertEquals(13335, data?.asn)
    }

    @Test
    fun testStandardEnvelopeErrorUnwrapping() = runTest {
        val errorEnvelope = """
            {
                "success": false,
                "data": null,
                "error": "rate limit exceeded on vantage point"
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String = errorEnvelope
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("1.1.1.1", timeoutSec = 5)

        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IOException)
        assertTrue(result.exceptionOrNull()?.message?.contains("rate limit exceeded") == true)
    }

    @Test
    fun testEnvelopeSuccessWithNullDataFailsSafely() = runTest {
        val corruptedEnvelope = """
            {
                "success": true,
                "data": null,
                "error": null
            }
        """.trimIndent()

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String = corruptedEnvelope
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("1.1.1.1", timeoutSec = 5)

        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IOException)
        assertTrue(result.exceptionOrNull()?.message?.contains("null data payload") == true)
    }

    @Test
    fun testEnvelopeErrorWithoutErrorMessageFailsSafely() = runTest {
        val errorEnvelopeNoMsg = """{"success": false}"""

        val mockDriver = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun queryBgpAsnJson(jsonRequest: String): String = errorEnvelopeNoMsg
        }

        val bridge = GoBridgeImpl(driver = mockDriver, ioDispatcher = Dispatchers.Default)
        val result = bridge.queryBgpAsn("1.1.1.1", timeoutSec = 5)

        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IOException)
        assertTrue(result.exceptionOrNull()?.message?.contains("without message") == true)
    }

    // ========================================================================
    // 2. Driver Linkage, Probing & Desktop Fallback Tests
    // ========================================================================

    @Test
    fun testJniDriverGracefulDegradationOnDesktop() {
        val jni = JniNativeBridgeDriver()
        // On desktop JVM, librastreador.so is not in java.library.path
        assertFalse("JniNativeBridgeDriver must report isAvailable() == false on desktop", jni.isAvailable())

        // Calling methods on unavailable driver must throw IllegalStateException, not crash native runtime
        assertThrows(IllegalStateException::class.java) {
            jni.queryBgpAsnJson("""{"ip": "1.1.1.1"}""")
        }
    }

    @Test
    fun testGomobileDriverGracefulDegradationOnDesktop() {
        val gomobile = GomobileBridgeDriver()
        // On desktop JVM without AAR in classpath, Mobile class is missing
        assertFalse("GomobileBridgeDriver must report isAvailable() == false on desktop", gomobile.isAvailable())

        assertThrows(IllegalStateException::class.java) {
            gomobile.queryBgpAsnJson("""{"ip": "1.1.1.1"}""")
        }
    }

    @Test
    fun testResolveDefaultDriverResolvesFallbackOnDesktop() {
        val driver = GoBridgeImpl.resolveDefaultDriver()
        assertNotNull(driver)
        assertTrue("Resolved driver must report available", driver.isAvailable())
        assertTrue(
            "Default driver on desktop JVM must be JvmFallbackBridgeDriver, got: ${driver::class.simpleName}",
            driver is JvmFallbackBridgeDriver
        )
    }

    @Test
    fun testWaterfallDriverResolutionPriority() {
        val unavailableJni = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun isAvailable(): Boolean = false
        }
        val availableGomobile = object : NativeBridgeDriver by JvmFallbackBridgeDriver() {
            override fun isAvailable(): Boolean = true
        }

        // When JNI is unavailable and Gomobile is available, driver should resolve to Gomobile
        val resolved = if (unavailableJni.isAvailable()) unavailableJni
        else if (availableGomobile.isAvailable()) availableGomobile
        else JvmFallbackBridgeDriver()

        assertSame(availableGomobile, resolved)
    }
}
