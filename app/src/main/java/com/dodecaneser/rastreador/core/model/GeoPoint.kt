package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class GeoPoint(
    @SerialName("lat") val lat: Double,
    @SerialName("lon") val lon: Double
)

typealias Point = GeoPoint
