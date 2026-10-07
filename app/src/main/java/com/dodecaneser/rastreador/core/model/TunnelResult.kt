package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class TunnelResult(
    @SerialName("target_ip") val targetIp: String,
    @SerialName("is_tunnel_detected") val isTunnelDetected: Boolean = false,
    @SerialName("tunnel_type") val tunnelType: String = "NONE", // NONE, WIREGUARD, OPENVPN, IPSEC, HTTP_PROXY
    @SerialName("mss_clamping_detected") val mssClampingDetected: Boolean = false,
    @SerialName("latency_inflation_ms") val latencyInflationMs: Double = 0.0,
    @SerialName("confidence") val confidence: String = "LOW",
    @SerialName("l4_tcp_rtt_ms") val l4TcpRttMs: Double = 0.0,
    @SerialName("l7_app_rtt_ms") val l7AppRttMs: Double = 0.0,
    @SerialName("delta_rtt_ms") val deltaRttMs: Double = 0.0,
    @SerialName("estimated_tunnel_km") val estimatedTunnelKm: Double = 0.0,
    @SerialName("findings") val findings: List<String> = emptyList()
)
