package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class WiFiBeaconScan(
    @SerialName("bssid") val bssid: String,
    @SerialName("ssid") val ssid: String = "",
    @SerialName("rssi") val rssi: Int = 0,
    @SerialName("frequency") val frequency: Int = 0,
    @SerialName("channel") val channel: Int = 0,
    @SerialName("channel_width") val channelWidth: Int = 20,
    @SerialName("capabilities") val capabilities: String = "[WPA2-PSK-CCMP][RSN]",
    @SerialName("latitude") val lat: Double = 0.0,
    @SerialName("longitude") val lon: Double = 0.0,
    @SerialName("timestamp") val timestamp: Long = System.currentTimeMillis()
)
