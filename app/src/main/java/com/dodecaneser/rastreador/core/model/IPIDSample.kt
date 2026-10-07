package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class IPIDSample(
    @SerialName("timestamp_ms") val timestampMs: Long = 0L,
    @SerialName("delta_ms") val deltaMs: Double = 0.0,
    @SerialName("ip_id") val ipId: Int = 0,
    @SerialName("delta_id") val deltaId: Int = 0,
    @SerialName("tcp_tsval") val tcpTsVal: Long = 0L,
    @SerialName("delta_tsval") val deltaTs: Long = 0L
)
