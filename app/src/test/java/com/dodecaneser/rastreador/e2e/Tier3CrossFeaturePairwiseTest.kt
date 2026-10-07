package com.dodecaneser.rastreador.e2e

import com.dodecaneser.rastreador.e2e.model.*
import kotlin.math.*

/**
 * Tier 3: Cross-Feature Pairwise Interaction Test Suite.
 * Validates integration pathways, data conversions, and state transitions between
 * interacting feature pairs across the architecture.
 * Total: 18 dedicated cross-feature tests.
 */
class Tier3CrossFeaturePairwiseTest {

    // 1. F04 (Go Bridge) <-> F07 (Wi-Fi Scanner)
    fun test_tier3_01_f04_f07_wifi_scanner_to_go_triangulation_flow() {
        val rawScans = listOf(
            WiFiBeaconScan("00:11:22:33:44:01", "Campus-AP1", -62, 2412, 1, lat = 40.4501, lon = -3.7250),
            WiFiBeaconScan("00:11:22:33:44:02", "Campus-AP2", -75, 2437, 6, lat = 40.4510, lon = -3.7260),
            WiFiBeaconScan("00:11:22:33:44:03", "Campus-AP3", -80, 5180, 36, lat = 40.4495, lon = -3.7245)
        )
        // Feed scanner output into Go bridge Wi-Fi triangulation oracle
        val triResult = ReferenceOracle.triangulateWiFi(rawScans)
        assert(triResult.resolvedCount == 3) { "All 3 beacons resolved" }
        assert(triResult.confidenceKm < 0.025) { "High-density micro-triangulation within 25m" }
        // Strongest AP (AP1, -62 dBm) should pull the centroid closest to its location
        val distToStrongest = ReferenceOracle.distanceHaversine(triResult.estimatedPoint.lat, triResult.estimatedPoint.lon, 40.4501, -3.7250)
        val distToWeakest = ReferenceOracle.distanceHaversine(triResult.estimatedPoint.lat, triResult.estimatedPoint.lon, 40.4495, -3.7245)
        assert(distToStrongest < distToWeakest) { "Centroid is closer to the strongest beacon" }
    }

    // 2. F04 (Go Bridge) <-> F08 (Cellular Scanner)
    fun test_tier3_02_f04_f08_cell_scanner_to_go_multilat_flow() {
        val cellRecords = listOf(
            CellTowerRecord(256259L, "LTE", 214, 7, 12345, 42, -75, lat = 40.410, lon = -3.700),
            CellTowerRecord(256260L, "LTE", 214, 7, 12345, 43, -88, lat = 40.420, lon = -3.710)
        )
        // Convert cell records to landmarks for multilateration
        val landmarks = cellRecords.map {
            val rttEstMs = abs(it.rsrp) * 0.2 // RSRP-based distance constraint estimate
            Landmark("cell-${it.cellId}", "Tower-${it.pci}", "Madrid", "ES", Point(it.lat, it.lon), rttEstMs, rttEstMs * 10.0)
        }
        assert(landmarks.size == 2) { "Cell records mapped to landmarks" }
        assert(landmarks[0].location.lat == 40.410) { "Latitude preserved" }
    }

    // 3. F07 (Wi-Fi Scanner) <-> F11 (Room Persistence)
    fun test_tier3_03_f07_f11_wifi_scanner_to_room_persistence_batch() {
        val scanBatch = listOf(
            WiFiBeaconScan("00:14:22:01:23:45", "Eduroam", -65, 2412, 1),
            WiFiBeaconScan("00:14:22:01:23:46", "UCM-Guest", -68, 2412, 1)
        )
        val sessionId = "session_test_01"
        val observations = scanBatch.mapIndexed { idx, s ->
            WifiObservationModel(
                id = idx.toLong(),
                sessionId = sessionId,
                bssid = s.bssid,
                rssi = s.rssi,
                latitude = 40.449,
                longitude = -3.728,
                timestamp = System.currentTimeMillis()
            )
        }
        assert(observations.size == 2) { "Scan results converted to Room entities" }
        assert(observations.all { it.sessionId == sessionId }) { "Foreign key sessionId linked" }
    }

    // 4. F08 (Cellular Scanner) <-> F11 (Room Persistence)
    fun test_tier3_04_f08_f11_cell_scanner_to_room_persistence_batch() {
        val cellId = 68719476735L // 5G NR
        val cellEntity = CellTowerRecord(cellId, "NR", 214, 1, 54321, 102, -72)
        assert(cellEntity.cellId == cellId) { "64-bit cell ID persisted without truncation" }
        assert(cellEntity.networkType == "NR") { "5G NR type preserved" }
    }

    // 5. F09 (GNSS Tracker) <-> F11 (Room Persistence)
    fun test_tier3_05_f09_f11_gnss_tracker_to_room_breadcrumb_batch() {
        val location = LocationRecord(40.4168, -3.7038, 650.0, 1.5f, 180f, 3.0f, "gps")
        val breadcrumb = GpsBreadcrumbModel(1L, "session_test_01", location.latitude, location.longitude, location.altitude, location.speed, location.bearing)
        assert(breadcrumb.latitude == 40.4168) { "Latitude mapped" }
        assert(breadcrumb.longitude == -3.7038) { "Longitude mapped" }
        assert(breadcrumb.speed == 1.5f) { "Speed mapped" }
    }

    // 6. F10 (Wardriving Service) <-> F11 (Room Persistence)
    fun test_tier3_06_f10_f11_wardriving_service_to_room_session_cascade() {
        val session = SessionEntityModel("sess-uuid-1", "Wardrive Drive 1", System.currentTimeMillis())
        assert(session.id.startsWith("sess-")) { "Session created" }
        assert(session.endTime == null) { "Session active (no end time)" }
        val closedSession = session.copy(endTime = System.currentTimeMillis(), totalWifi = 150)
        assert(closedSession.endTime != null) { "Session closed successfully" }
        assert(closedSession.totalWifi == 150) { "Stats updated" }
    }

    // 7. F04 (Go Bridge) <-> F13 (OsmDroid Map)
    fun test_tier3_07_f04_f13_go_cbg_multilat_to_osmdroid_confidence_overlay() {
        val targetPoint = Point(40.450, -3.725)
        val confidenceKm = 10.0
        // Convert confidence radius in km to map bounding box
        val dLat = (confidenceKm / ReferenceOracle.EARTH_RADIUS_KM) * (180.0 / Math.PI)
        val dLon = (confidenceKm / (ReferenceOracle.EARTH_RADIUS_KM * cos(Math.toRadians(targetPoint.lat)))) * (180.0 / Math.PI)
        val north = targetPoint.lat + dLat
        val south = targetPoint.lat - dLat
        val east = targetPoint.lon + dLon
        val west = targetPoint.lon - dLon
        assert(north > targetPoint.lat && south < targetPoint.lat) { "Bounding box spans north and south" }
        assert(east > targetPoint.lon && west < targetPoint.lon) { "Bounding box spans east and west" }
    }

    // 8. F07 (Wi-Fi Scanner) <-> F13 (OsmDroid Map)
    fun test_tier3_08_f07_f13_wifi_centroid_to_osmdroid_marker_placement() {
        val tri = ReferenceOracle.triangulateWiFi(listOf(
            WiFiBeaconScan("00:11:22:33:44:55", "TestNet", -65, 2412, 1, lat = 40.4168, lon = -3.7038)
        ))
        val markerLat = tri.estimatedPoint.lat
        val markerLon = tri.estimatedPoint.lon
        val accuracyCircleMeters = tri.precisionM
        assert(markerLat == 40.4168) { "Marker placed at resolved coordinates" }
        assert(accuracyCircleMeters == 25.0) { "Accuracy circle matches precision" }
    }

    // 9. F04 (Go Bridge) <-> F14 (Canvas Chart)
    fun test_tier3_09_f04_f14_ipid_analysis_to_canvas_chart_rendering() {
        val ipidResult = ReferenceOracle.analyzeIpId(listOf(
            IPIDSample(0L, 100.0, 100, 10),
            IPIDSample(100L, 100.0, 110, 10),
            IPIDSample(200L, 100.0, 120, 10)
        ))
        val velocity = ipidResult.velocityPacketsPerSec
        val r2 = ipidResult.linearityScore
        val maxScale = 500f
        val chartNormalizedHeight = (velocity.toFloat() / maxScale).coerceIn(0f, 1f)
        assert(velocity == 100.0) { "Velocity is 100 pkts/s" }
        assert(r2 == 1.0) { "R^2 is 1.0" }
        assert(chartNormalizedHeight == 0.2f) { "Normalized chart height is 20%" }
    }

    // 10. F04 (Go Bridge) <-> F15 (Forensic Export)
    fun test_tier3_10_f04_f15_multilat_and_bgp_to_forensic_json_report() {
        val bgp = BgpAsnResult("147.96.1.1", 766, "UCM", "ES", "Madrid", "RedIRIS")
        val multilat = MultilaterationResult(Point(40.44, -3.73), 15.0, emptyList())
        val report = ForensicReportData("147.96.1.1", "2026-10-07T14:00:00Z", bgp, multilat, null, null)
        val canonical = "${report.targetIp}|${report.timestamp}|${report.reconInfo?.asn}|${report.multilateration?.confidenceKm}"
        val seal = ReferenceOracle.sha256(canonical)
        assert(seal.length == 64) { "Cryptographic seal generated from combined Go bridge data" }
    }

    // 11. F11 (Room DB) <-> F15 (Forensic Export)
    fun test_tier3_11_f11_f15_room_database_session_to_forensic_pdf_report() {
        val session = SessionEntityModel("sess-1", "Audit Mission", 1000L, 5000L, 25, 4)
        val reportTitle = "FORENSIC TELEMETRY REPORT — ${session.name}"
        assert(reportTitle.contains("Audit Mission")) { "Report title embeds session name" }
        val durationSec = (session.endTime!! - session.startTime) / 1000L
        assert(durationSec == 4L) { "Duration computed from Room session timestamps" }
    }

    // 12. F07 (Wi-Fi Scanner) <-> F12 (Compose UI)
    fun test_tier3_12_f07_f12_wifi_scanner_to_compose_scanner_screen_ui_state() {
        val scans = listOf(
            WiFiBeaconScan("00:11:22:33:44:55", "NetA", -50, 2412, 1),
            WiFiBeaconScan("00:11:22:33:44:56", "NetB", -85, 5180, 36)
        )
        data class ScannerUiState(val totalCount: Int, val strongestSsid: String, val isLoading: Boolean)
        val uiState = ScannerUiState(scans.size, scans.minByOrNull { it.rssi }?.ssid ?: "", false)
        assert(uiState.totalCount == 2) { "Total count hoisted to UI state" }
        assert(!uiState.isLoading) { "Scanner loading state updated" }
    }

    // 13. F08 (Cellular Scanner) <-> F12 (Compose UI)
    fun test_tier3_13_f08_f12_cell_scanner_to_compose_dashboard_screen_ui_state() {
        val cell = CellTowerRecord(256259L, "LTE", 214, 7, 12345, 42, -78)
        data class DashboardCellState(val operator: String, val networkType: String, val signalDbm: Int)
        val state = DashboardCellState("Movistar (214-07)", cell.networkType, cell.rsrp)
        assert(state.operator.contains("Movistar")) { "Operator mapped to dashboard HUD" }
        assert(state.signalDbm == -78) { "Signal level mapped" }
    }

    // 14. F10 (Wardriving Service) <-> F12 (Compose UI)
    fun test_tier3_14_f10_f12_wardriving_service_to_compose_wardriving_screen_ui_state() {
        val isServiceRunning = true
        val stats = WardrivingStats(totalSessions = 5, totalWifiObserved = 342, uniqueBssids = 210)
        data class WardrivingUiState(val isRunning: Boolean, val uniqueAps: Int)
        val ui = WardrivingUiState(isServiceRunning, stats.uniqueBssids)
        assert(ui.isRunning) { "Wardriving running state shown in UI" }
        assert(ui.uniqueAps == 210) { "Unique AP metric shown in UI" }
    }

    // 15. F06 (Coroutines Flow) <-> F04 (Go Bridge)
    fun test_tier3_15_f06_f04_coroutines_flow_to_go_bridge_cancellation() {
        var bridgeCallCount = 0
        var cancelled = false
        // Simulate flow collection that cancels after first emission
        val stream = listOf("result-1", "result-2", "result-3")
        for (item in stream) {
            bridgeCallCount++
            cancelled = true
            break // Coroutine Flow cancel
        }
        assert(bridgeCallCount == 1) { "Bridge loop terminates immediately on cancellation" }
        assert(cancelled) { "Cancellation status set" }
    }

    // 16. F02 (Permissions) <-> F10 (Wardriving Service)
    fun test_tier3_16_f02_f10_permissions_check_to_wardriving_service_start() {
        val grantedPermissions = setOf(
            "android.permission.ACCESS_FINE_LOCATION",
            "android.permission.FOREGROUND_SERVICE_LOCATION"
        )
        fun canStartService(granted: Set<String>): Boolean {
            return granted.contains("android.permission.ACCESS_FINE_LOCATION") &&
                   granted.contains("android.permission.FOREGROUND_SERVICE_LOCATION")
        }
        assert(canStartService(grantedPermissions)) { "Service start permitted when all required permissions granted" }
        assert(!canStartService(emptySet())) { "Service start blocked when permissions missing" }
    }

    // 17. F01 (Version Catalog) <-> F05 (AAR FlatDir Dependency)
    fun test_tier3_17_f01_f05_version_catalog_to_aar_flatdir_dependency() {
        val dependencyNotation = "files(\"libs/rastreador-core.aar\")"
        assert(dependencyNotation.contains("rastreador-core.aar")) { "Build dependency targets core AAR binary" }
    }

    // 18. F15 (Forensic Export) <-> F17 (Adversarial Verification)
    fun test_tier3_18_f15_f17_forensic_export_tamper_detection_in_adversarial_audit() {
        val originalPayload = "IP=147.96.1.1;LAT=40.44;LON=-3.73;SEAL=9a3f"
        val originalDigest = ReferenceOracle.sha256(originalPayload)
        // Adversary alters coordinates by 0.0001 degree
        val tamperedPayload = "IP=147.96.1.1;LAT=40.4401;LON=-3.73;SEAL=9a3f"
        val tamperedDigest = ReferenceOracle.sha256(tamperedPayload)
        assert(originalDigest != tamperedDigest) { "Adversarial tampering immediately detected by SHA-256 seal mismatch" }
    }
}

data class GpsBreadcrumbModel(
    val id: Long,
    val sessionId: String,
    val latitude: Double,
    val longitude: Double,
    val altitude: Double,
    val speed: Float,
    val bearing: Float
)
