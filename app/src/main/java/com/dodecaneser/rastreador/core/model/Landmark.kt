package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Landmark(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String,
    @SerialName("city") val city: String,
    @SerialName("country") val country: String,
    @SerialName("location") val location: GeoPoint,
    @SerialName("min_rtt_ms") val minRttMs: Double,
    @SerialName("max_radius_km") val maxRadiusKm: Double,
    @SerialName("samples") val samples: Int = 3,
    @SerialName("type") val type: String = "hop_landmark"
)
