package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.serialization.SerializationException
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Test
import java.io.IOException

class DomainModelSerializationTest {

    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        coerceInputValues = true
        encodeDefaults = true
    }

    @Test
    fun testBgpAsnResultSerialization() {
        val original = BgpAsnResult(
            ip = "147.96.1.1",
            asn = 766,
            asOrg = "Universidad Complutense de Madrid",
            country = "ES",
            countryCode = "ES",
            city = "Madrid",
            region = "Comunidad de Madrid",
            zip = "28040",
            facility = "Ciudad Universitaria Campus",
            latitude = 40.4489,
            longitude = -3.7297,
            isp = "RedIRIS",
            hostname = "www.ucm.es",
            isCloud = false,
            isVpn = false,
            isProxy = false,
            isTorExit = false,
            isAnycast = false,
            confidence = "HIGH",
            precisionKm = 0.5,
            airportCode = "MAD",
            indicators = listOf("PTR: www.ucm.es", "Facility match")
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<BgpAsnResult>(serialized)

        assertEquals(original.ip, deserialized.ip)
        assertEquals(original.asn, deserialized.asn)
        assertEquals(original.asOrg, deserialized.asOrg)
        assertEquals(original.country, deserialized.country)
        assertEquals(original.city, deserialized.city)
        assertEquals(original.facility, deserialized.facility)
        assertEquals(original.latitude, deserialized.latitude, 0.0001)
        assertEquals(original.longitude, deserialized.longitude, 0.0001)
        assertEquals(original.confidence, deserialized.confidence)
        assertEquals(2, deserialized.indicators.size)
    }

    @Test
    fun testBgpAsnResultWithUnknownKeysForwardCompatibility() {
        val jsonWithExtras = """
            {
                "ip": "8.8.8.8",
                "asn": 15169,
                "as_org": "Google LLC",
                "country": "US",
                "city": "Mountain View",
                "isp": "Google",
                "future_experimental_field": "test_value_123",
                "internal_debug_flags": 9999
            }
        """.trimIndent()

        val parsed = json.decodeFromString<BgpAsnResult>(jsonWithExtras)
        assertEquals("8.8.8.8", parsed.ip)
        assertEquals(15169, parsed.asn)
        assertEquals("Google LLC", parsed.asOrg)
        assertEquals("US", parsed.country)
        assertEquals("Mountain View", parsed.city)
        // Default preserved
        assertEquals("HIGH", parsed.confidence)
    }

    @Test
    fun testBgpAsnResultMissingOptionalFields() {
        val minimalJson = """{"ip": "1.1.1.1"}"""
        val parsed = json.decodeFromString<BgpAsnResult>(minimalJson)

        assertEquals("1.1.1.1", parsed.ip)
        assertEquals(0, parsed.asn)
        assertEquals("", parsed.asOrg)
        assertFalse(parsed.isVpn)
        assertFalse(parsed.isCloud)
        assertTrue(parsed.indicators.isEmpty())
    }

    @Test
    fun testMultilaterationResultSerialization() {
        val landmarks = listOf(
            Landmark("lm-mad", "Madrid Hub", "Madrid", "ES", GeoPoint(40.4168, -3.7038), 2.5, 250.0),
            Landmark("lm-bcn", "Barcelona Hub", "Barcelona", "ES", GeoPoint(41.3851, 2.1734), 8.0, 800.0)
        )
        val original = MultilaterationResult(
            estimatedPoint = GeoPoint(40.4489, -3.7270),
            confidenceKm = 8.5,
            usedLandmarks = landmarks,
            polygonBounds = listOf(GeoPoint(40.0, -4.0), GeoPoint(41.0, -3.0))
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MultilaterationResult>(serialized)

        assertEquals(original.estimatedPoint.lat, deserialized.estimatedPoint.lat, 0.0001)
        assertEquals(original.estimatedPoint.lon, deserialized.estimatedPoint.lon, 0.0001)
        assertEquals(original.confidenceKm, deserialized.confidenceKm, 0.0001)
        assertEquals(2, deserialized.usedLandmarks.size)
        assertEquals("Madrid Hub", deserialized.usedLandmarks[0].name)
        assertEquals(2, deserialized.polygonBounds.size)
    }

    @Test
    fun testTriangulationResultSerialization() {
        val beacons = listOf(
            WiFiBeaconScan("00:11:22:33:44:55", "Office-Wifi", -55, 2412, 1, 20, "[WPA2]", 40.4168, -3.7038),
            WiFiBeaconScan("00:11:22:33:44:66", "Guest-Wifi", -65, 5180, 36, 40, "[WPA3]", 40.4170, -3.7040)
        )
        val original = TriangulationResult(
            estimatedPoint = GeoPoint(40.4169, -3.7039),
            confidenceKm = 0.015,
            precisionM = 15.0,
            resolvedCount = 2,
            totalBeacons = 2,
            streetAddress = "Calle Gran Vía 1, Madrid",
            networks = beacons
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<TriangulationResult>(serialized)

        assertEquals(original.estimatedPoint.lat, deserialized.estimatedPoint.lat, 0.0001)
        assertEquals(15.0, deserialized.precisionM, 0.001)
        assertEquals(2, deserialized.resolvedCount)
        assertEquals("Calle Gran Vía 1, Madrid", deserialized.streetAddress)
        assertEquals(2, deserialized.networks.size)
        assertEquals("Office-Wifi", deserialized.networks[0].ssid)
    }

    @Test
    fun testIpIdResultSerialization() {
        val samples = listOf(
            IPIDSample(1000L, 100.0, 1000, 10, 10000L, 1000L),
            IPIDSample(1100L, 100.0, 1010, 10, 11000L, 1000L)
        )
        val original = IpIdResult(
            targetIp = "147.96.1.1",
            timestamp = "2026-10-07T00:00:00Z",
            generationType = "GLOBAL_INCREMENTAL",
            velocityPacketsPerSec = 100.0,
            clockFrequencyHz = 250.0,
            linearityScore = 0.995,
            correlationFingerprint = "a1b2c3d4e5f60718",
            totalSamples = 2,
            samples = samples,
            findings = listOf("Linear increment detected")
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<IpIdResult>(serialized)

        assertEquals("147.96.1.1", deserialized.targetIp)
        assertEquals("GLOBAL_INCREMENTAL", deserialized.generationType)
        assertEquals(100.0, deserialized.velocityPacketsPerSec, 0.001)
        assertEquals(0.995, deserialized.linearityScore, 0.001)
        assertEquals(2, deserialized.samples.size)
        assertEquals(1010, deserialized.samples[1].ipId)
    }

    @Test
    fun testTunnelResultSerialization() {
        val original = TunnelResult(
            targetIp = "10.0.0.1",
            isTunnelDetected = true,
            tunnelType = "WIREGUARD",
            mssClampingDetected = true,
            latencyInflationMs = 42.0,
            confidence = "HIGH",
            l4TcpRttMs = 15.0,
            l7AppRttMs = 57.0,
            deltaRttMs = 42.0,
            estimatedTunnelKm = 4200.0,
            findings = listOf("High latency delta", "MSS clamped")
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<TunnelResult>(serialized)

        assertEquals("10.0.0.1", deserialized.targetIp)
        assertTrue(deserialized.isTunnelDetected)
        assertEquals("WIREGUARD", deserialized.tunnelType)
        assertTrue(deserialized.mssClampingDetected)
        assertEquals(42.0, deserialized.latencyInflationMs, 0.001)
        assertEquals("HIGH", deserialized.confidence)
        assertEquals(2, deserialized.findings.size)
    }

    @Test
    fun testBridgeEnvelopeSuccessUnwrapping() {
        val envelopeJson = """
            {
                "success": true,
                "data": {
                    "ip": "1.1.1.1",
                    "asn": 13335,
                    "as_org": "Cloudflare, Inc.",
                    "country": "US",
                    "city": "San Francisco",
                    "isp": "Cloudflare",
                    "is_anycast": true
                },
                "error": null
            }
        """.trimIndent()

        val envelope = json.decodeFromString<BridgeEnvelope<BgpAsnResult>>(envelopeJson)
        assertTrue(envelope.success)
        assertNotNull(envelope.data)
        assertNull(envelope.error)
        assertEquals(13335, envelope.data?.asn)
        assertTrue(envelope.data?.isAnycast == true)
    }

    @Test
    fun testBridgeEnvelopeErrorHandling() {
        val envelopeJson = """
            {
                "success": false,
                "error": "invalid IP address: 999.999.999.999"
            }
        """.trimIndent()

        val envelope = json.decodeFromString<BridgeEnvelope<BgpAsnResult>>(envelopeJson)
        assertFalse(envelope.success)
        assertNull(envelope.data)
        assertEquals("invalid IP address: 999.999.999.999", envelope.error)
    }

    @Test
    fun testMalformedJsonRejection() {
        assertThrows(SerializationException::class.java) {
            json.decodeFromString<BgpAsnResult>("NOT_VALID_JSON_AT_ALL")
        }
    }
}
