package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class BgpAsnResult(
    @SerialName("ip") val ip: String,
    @SerialName("asn") val asn: Int = 0,
    @SerialName("as_org") val asOrg: String = "",
    @SerialName("country") val country: String = "",
    @SerialName("country_code") val countryCode: String = "",
    @SerialName("city") val city: String = "",
    @SerialName("region") val region: String = "",
    @SerialName("zip") val zip: String = "",
    @SerialName("facility") val facility: String = "",
    @SerialName("latitude") val latitude: Double = 0.0,
    @SerialName("longitude") val longitude: Double = 0.0,
    @SerialName("isp") val isp: String = "",
    @SerialName("hostname") val hostname: String = "",
    @SerialName("is_cloud") val isCloud: Boolean = false,
    @SerialName("is_vpn") val isVpn: Boolean = false,
    @SerialName("is_proxy") val isProxy: Boolean = false,
    @SerialName("is_tor_exit") val isTorExit: Boolean = false,
    @SerialName("is_anycast") val isAnycast: Boolean = false,
    @SerialName("confidence") val confidence: String = "HIGH",
    @SerialName("precision_km") val precisionKm: Double = 0.0,
    @SerialName("detected_airport_code") val airportCode: String = "",
    @SerialName("indicators") val indicators: List<String> = emptyList()
)
