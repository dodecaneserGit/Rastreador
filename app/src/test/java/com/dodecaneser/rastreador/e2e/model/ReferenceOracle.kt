package com.dodecaneser.rastreador.e2e.model

import java.security.MessageDigest
import kotlin.math.*

/**
 * Authoritative reference implementations (oracles) derived directly from:
 * 1. Go core engine at /Volumes/SSD/Proyectos/Rastreador (pkg/multilat/geo.go, pkg/l2wifi/locator.go, pkg/ipid/analyzer.go)
 * 2. 3GPP LTE / 5G NR specs (TS 36.331, TS 38.331)
 * 3. RFC 6864, RFC 1323
 */
object ReferenceOracle {

    const val EARTH_RADIUS_KM = 6371.0
    const val SPEED_OF_LIGHT_FIBER_KM_PER_MS = 200.0

    /**
     * Great-circle distance between two geographic coordinates using the Haversine formula.
     * Matches Go implementation in pkg/multilat/geo.go DistanceHaversine.
     */
    fun distanceHaversine(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
        val dLat = Math.toRadians(lat2 - lat1)
        val dLon = Math.toRadians(lon2 - lon1)
        val rLat1 = Math.toRadians(lat1)
        val rLat2 = Math.toRadians(lat2)

        val a = sin(dLat / 2.0).pow(2) +
                sin(dLon / 2.0).pow(2) * cos(rLat1) * cos(rLat2)
        val c = 2.0 * atan2(sqrt(a), sqrt(1.0 - a))
        return EARTH_RADIUS_KM * c
    }

    /**
     * Constraint radius in kilometers from round-trip time in milliseconds (CBG model).
     * Matches Go implementation in pkg/multilat/geo.go ConstraintRadiusFromRTT.
     */
    fun constraintRadiusFromRTT(rttMs: Double): Double {
        val oneWayMs = rttMs / 2.0
        return oneWayMs * SPEED_OF_LIGHT_FIBER_KM_PER_MS
    }

    /**
     * Triangulate Wi-Fi beacons using RSSI log-distance path loss weighted centroid.
     * Matches Go implementation in pkg/l2wifi/locator.go TriangulateBSSIDs / TriangulateWiFi.
     */
    fun triangulateWiFi(beacons: List<WiFiBeaconScan>): TriangulationResult {
        val validNets = beacons.filter { it.lat != 0.0 && it.lon != 0.0 }
        if (validNets.isEmpty()) {
            return TriangulationResult(
                estimatedPoint = Point(0.0, 0.0),
                confidenceKm = 50.0,
                precisionM = 50000.0,
                resolvedCount = 0,
                totalBeacons = beacons.size
            )
        }

        if (validNets.size == 1) {
            val single = validNets[0]
            return TriangulationResult(
                estimatedPoint = Point(single.lat, single.lon),
                confidenceKm = 0.025,
                precisionM = 25.0,
                resolvedCount = 1,
                totalBeacons = beacons.size,
                networks = validNets
            )
        }

        var sumLat = 0.0
        var sumLon = 0.0
        var sumWeight = 0.0

        for (b in validNets) {
            val rssi = if (b.rssi == 0) -70 else b.rssi
            // Weight w = 10^(RSSI / 20)
            var weight = 10.0.pow(rssi.toDouble() / 20.0)
            if (weight <= 0.0) weight = 0.001

            sumLat += b.lat * weight
            sumLon += b.lon * weight
            sumWeight += weight
        }

        val estLat = sumLat / sumWeight
        val estLon = sumLon / sumWeight
        val estPt = Point(estLat, estLon)

        var maxDistKm = 0.0
        for (b in validNets) {
            val d = distanceHaversine(estPt.lat, estPt.lon, b.lat, b.lon)
            if (d > maxDistKm) {
                maxDistKm = d
            }
        }

        val precM = max(8.0, min(25.0, maxDistKm * 1000.0 * 0.6))
        val confidenceKm = precM / 1000.0

        return TriangulationResult(
            estimatedPoint = estPt,
            confidenceKm = confidenceKm,
            precisionM = round(precM * 10.0) / 10.0,
            resolvedCount = validNets.size,
            totalBeacons = beacons.size,
            networks = validNets
        )
    }

    /**
     * Statistical linear regression analysis of IP-ID packet velocity and clocking.
     * Matches Go implementation in pkg/ipid/analyzer.go calculateVelocityAndLinearity.
     */
    fun analyzeIpId(samples: List<IPIDSample>, targetIp: String = "127.0.0.1"): IpIdResult {
        if (samples.size < 3) {
            return IpIdResult(
                targetIp = targetIp,
                timestamp = "2026-10-07T00:00:00Z",
                generationType = "INSUFFICIENT_SAMPLES",
                velocityPacketsPerSec = 0.0,
                linearityScore = 0.0,
                correlationFingerprint = "0000000000000000",
                totalSamples = samples.size,
                findings = listOf("Insufficient samples")
            )
        }

        // Check if constant zero (RFC 6864)
        val isZero = samples.all { it.ipId == 0 }
        if (isZero) {
            return IpIdResult(
                targetIp = targetIp,
                timestamp = "2026-10-07T00:00:00Z",
                generationType = "CONSTANT_ZERO",
                velocityPacketsPerSec = 0.0,
                linearityScore = 1.0,
                correlationFingerprint = computeFingerprint("CONSTANT_ZERO", targetIp, 0.0, 0.0),
                totalSamples = samples.size,
                findings = listOf("Constant zero IP-ID RFC 6864")
            )
        }

        val velocities = mutableListOf<Double>()
        for (s in samples) {
            if (s.deltaMs > 0.0) {
                val v = s.deltaId / (s.deltaMs / 1000.0)
                velocities.add(v)
            }
        }

        val meanVel = if (velocities.isNotEmpty()) velocities.average() else 0.0
        val sumSqDiff = velocities.sumOf { (it - meanVel).pow(2) }
        val stdDev = if (velocities.isNotEmpty()) sqrt(sumSqDiff / velocities.size) else 0.0

        var r2 = 0.0
        if (meanVel > 0.0) {
            val cv = stdDev / meanVel
            r2 = max(0.0, min(1.0, 1.0 - (cv * 0.5)))
        }

        val roundedVel = round(meanVel * 100.0) / 100.0
        val roundedR2 = round(r2 * 1000.0) / 1000.0

        val genType = when {
            r2 >= 0.85 && stdDev < (meanVel * 0.6) -> "GLOBAL_INCREMENTAL"
            stdDev > (meanVel * 2.0) -> "RANDOMIZED"
            else -> "PER_HOST_HASH"
        }

        val fp = computeFingerprint(genType, targetIp, roundedVel, 0.0)

        return IpIdResult(
            targetIp = targetIp,
            timestamp = "2026-10-07T00:00:00Z",
            generationType = genType,
            velocityPacketsPerSec = roundedVel,
            linearityScore = roundedR2,
            correlationFingerprint = fp,
            totalSamples = samples.size,
            samples = samples
        )
    }

    /**
     * Compute 16-character hexadecimal correlation fingerprint from target metrics.
     */
    fun computeFingerprint(genType: String, ip: String, vel: Double, clockHz: Double): String {
        val input = String.format("%s:%s:%.1f:%.0f", genType, ip, vel, clockHz)
        val hash = sha256(input)
        return hash.substring(0, min(16, hash.length))
    }

    /**
     * SHA-256 cryptographic digest matching standard NIST FIPS 180-4.
     */
    fun sha256(input: String): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val bytes = digest.digest(input.toByteArray(Charsets.UTF_8))
        return bytes.joinToString("") { "%02x".format(it) }
    }

    /**
     * Parse 28-bit LTE Cell Identity (CI) into 20-bit eNodeB ID and 8-bit Sector ID.
     * TS 36.331 specification.
     */
    fun parseLteCellId(ci: Int, tac: Int = 0, pci: Int = 0, mcc: Int = 214, mnc: Int = 1): LteCellIdentity {
        val eNodeB = ci ushr 8
        val sectorId = ci and 0xFF
        return LteCellIdentity(
            ci = ci,
            eNodeB = eNodeB,
            sectorId = sectorId,
            tac = tac,
            pci = pci,
            mcc = mcc,
            mnc = mnc
        )
    }

    /**
     * Parse 36-bit 5G NR Cell Identity (NCI) into 22-bit gNodeB ID and 14-bit Sector ID.
     * TS 38.331 specification.
     */
    fun parseNrCellId(nci: Long, tac: Int = 0, pci: Int = 0, mcc: Int = 214, mnc: Int = 1): NrCellIdentity {
        val gNodeB = nci ushr 14
        val sectorId = (nci and 0x3FFF).toInt()
        return NrCellIdentity(
            nci = nci,
            gNodeB = gNodeB,
            sectorId = sectorId,
            tac = tac,
            pci = pci,
            mcc = mcc,
            mnc = mnc
        )
    }

    /**
     * Map Wi-Fi operating frequency in MHz to standard 802.11 channel number.
     */
    fun frequencyToChannel(freqMHz: Int): Int {
        return when {
            freqMHz in 2412..2472 -> (freqMHz - 2407) / 5
            freqMHz == 2484 -> 14
            freqMHz in 5170..5825 -> (freqMHz - 5000) / 5
            freqMHz in 5935..7115 -> (freqMHz - 5950) / 5
            else -> 0
        }
    }

    /**
     * Normalize BSSID string into standard lowercase colon-delimited format: 00:11:22:33:44:55
     */
    fun normalizeBssid(raw: String): String {
        val cleaned = raw.replace("[-.:]".toRegex(), "").lowercase()
        if (cleaned.length != 12) return raw.lowercase()
        return cleaned.chunked(2).joinToString(":")
    }
}
