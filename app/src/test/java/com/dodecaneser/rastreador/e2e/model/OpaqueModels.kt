package com.dodecaneser.rastreador.e2e.model

/**
 * Opaque-box data models and interface definitions representing the Rastreador Mobile architecture.
 * Derived from ORIGINAL_REQUEST.md, PROJECT.md, and system specifications.
 */

data class Point(
    val lat: Double,
    val lon: Double
)

data class Landmark(
    val id: String,
    val name: String,
    val city: String,
    val country: String,
    val location: Point,
    val minRttMs: Double,
    val maxRadiusKm: Double,
    val samples: Int = 3,
    val type: String = "hop_landmark"
)

data class MultilaterationResult(
    val estimatedPoint: Point,
    val confidenceKm: Double,
    val usedLandmarks: List<Landmark>,
    val polygonBounds: List<Point> = emptyList()
)

data class BgpAsnResult(
    val ip: String,
    val asn: Int,
    val asOrg: String,
    val country: String,
    val city: String,
    val isp: String,
    val isVpn: Boolean = false,
    val isProxy: Boolean = false,
    val isCloud: Boolean = false,
    val confidence: String = "HIGH"
)

data class WiFiBeaconScan(
    val bssid: String,
    val ssid: String,
    val rssi: Int,
    val frequency: Int,
    val channel: Int,
    val channelWidth: Int = 20,
    val capabilities: String = "[WPA2-PSK-CCMP][RSN]",
    val lat: Double = 0.0,
    val lon: Double = 0.0,
    val timestamp: Long = System.currentTimeMillis()
)

data class TriangulationResult(
    val estimatedPoint: Point,
    val confidenceKm: Double,
    val precisionM: Double,
    val resolvedCount: Int,
    val totalBeacons: Int,
    val streetAddress: String = "",
    val networks: List<WiFiBeaconScan> = emptyList()
)

data class IPIDSample(
    val timestampMs: Long,
    val deltaMs: Double,
    val ipId: Int,
    val deltaId: Int,
    val tcpTsVal: Long = 0L,
    val deltaTs: Long = 0L
)

data class IpIdResult(
    val targetIp: String,
    val timestamp: String,
    val generationType: String, // "GLOBAL_INCREMENTAL", "RANDOMIZED", "CONSTANT_ZERO", "PER_HOST_HASH"
    val velocityPacketsPerSec: Double,
    val clockFrequencyHz: Double = 0.0,
    val linearityScore: Double, // R^2 score
    val correlationFingerprint: String,
    val totalSamples: Int,
    val samples: List<IPIDSample> = emptyList(),
    val findings: List<String> = emptyList()
)

data class TunnelResult(
    val targetIp: String,
    val isTunnelDetected: Boolean,
    val tunnelType: String, // "NONE", "WIREGUARD", "OPENVPN", "IPSEC", "HTTP_PROXY"
    val mssClampingDetected: Boolean,
    val latencyInflationMs: Double,
    val confidence: String
)

data class LteCellIdentity(
    val ci: Int, // 28-bit
    val eNodeB: Int, // 20-bit
    val sectorId: Int, // 8-bit
    val tac: Int,
    val pci: Int,
    val mcc: Int,
    val mnc: Int
)

data class NrCellIdentity(
    val nci: Long, // 36-bit Long
    val gNodeB: Long, // 22-bit
    val sectorId: Int, // 14-bit
    val tac: Int,
    val pci: Int,
    val mcc: Int,
    val mnc: Int
)

data class CellTowerRecord(
    val cellId: Long,
    val networkType: String, // "LTE", "NR", "GSM", "WCDMA"
    val mcc: Int,
    val mnc: Int,
    val tacLac: Int,
    val pci: Int,
    val rsrp: Int,
    val lat: Double = 0.0,
    val lon: Double = 0.0,
    val isRegistered: Boolean = true
)

data class LocationRecord(
    val latitude: Double,
    val longitude: Double,
    val altitude: Double,
    val speed: Float,
    val bearing: Float,
    val accuracy: Float,
    val provider: String, // "fused" or "gps"
    val timestamp: Long = System.currentTimeMillis()
)

data class WardrivingStats(
    val totalSessions: Int = 0,
    val totalWifiObserved: Int = 0,
    val uniqueBssids: Int = 0,
    val totalCellsObserved: Int = 0,
    val uniqueCells: Int = 0,
    val totalBreadcrumbs: Int = 0,
    val distanceTraveledKm: Double = 0.0
)

data class SessionEntityModel(
    val id: String,
    val name: String,
    val startTime: Long,
    val endTime: Long? = null,
    val totalWifi: Int = 0,
    val totalCells: Int = 0
)

data class WifiObservationModel(
    val id: Long = 0,
    val sessionId: String,
    val bssid: String,
    val rssi: Int,
    val latitude: Double,
    val longitude: Double,
    val timestamp: Long
)

data class ForensicReportData(
    val targetIp: String,
    val timestamp: String,
    val reconInfo: BgpAsnResult?,
    val multilateration: MultilaterationResult?,
    val ipidAnalysis: IpIdResult?,
    val wifiTriangulation: TriangulationResult?,
    val cellularTelemetry: List<CellTowerRecord> = emptyList(),
    val gnssTelemetry: List<LocationRecord> = emptyList()
)

enum class MapTileLayer {
    ESRI_SATELLITE,
    CARTO_DARK,
    OPEN_TOPO
}
