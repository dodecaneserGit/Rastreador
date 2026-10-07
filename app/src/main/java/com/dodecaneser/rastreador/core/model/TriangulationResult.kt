package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class TriangulationResult(
    @SerialName("estimated_point") val estimatedPoint: GeoPoint,
    @SerialName("confidence_km") val confidenceKm: Double,
    @SerialName("precision_meters") val precisionM: Double = 0.0,
    @SerialName("resolved_count") val resolvedCount: Int = 0,
    @SerialName("total_beacons") val totalBeacons: Int = 0,
    @SerialName("street_address") val streetAddress: String = "",
    @SerialName("networks") val networks: List<WiFiBeaconScan> = emptyList()
)
