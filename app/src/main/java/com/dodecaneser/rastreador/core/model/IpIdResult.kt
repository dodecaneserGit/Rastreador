package com.dodecaneser.rastreador.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class IpIdResult(
    @SerialName("target_ip") val targetIp: String,
    @SerialName("timestamp") val timestamp: String = "",
    @SerialName("generation_type") val generationType: String, // GLOBAL_INCREMENTAL, RANDOMIZED, CONSTANT_ZERO, PER_HOST_HASH
    @SerialName("velocity_packets_per_sec") val velocityPacketsPerSec: Double,
    @SerialName("clock_frequency_hz") val clockFrequencyHz: Double = 0.0,
    @SerialName("linearity_score") val linearityScore: Double, // R^2
    @SerialName("correlation_fingerprint") val correlationFingerprint: String,
    @SerialName("total_samples") val totalSamples: Int = 0,
    @SerialName("samples") val samples: List<IPIDSample> = emptyList(),
    @SerialName("findings") val findings: List<String> = emptyList()
)
