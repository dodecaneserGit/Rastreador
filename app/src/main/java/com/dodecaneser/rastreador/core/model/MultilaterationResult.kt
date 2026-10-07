package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class MultilaterationResult(
    @SerialName("estimated_point") val estimatedPoint: GeoPoint,
    @SerialName("confidence_km") val confidenceKm: Double,
    @SerialName("used_landmarks") val usedLandmarks: List<Landmark> = emptyList(),
    @SerialName("polygon_bounds") val polygonBounds: List<GeoPoint> = emptyList()
)
