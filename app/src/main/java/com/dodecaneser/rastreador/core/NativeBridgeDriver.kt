package com.dodecaneser.rastreador.core

/**
 * Low-level driver interface abstracting JNI native calls from pure JVM fallback execution.
 */
interface NativeBridgeDriver {
    fun isAvailable(): Boolean
    fun queryBgpAsnJson(jsonRequest: String): String
    fun performMultilaterationJson(jsonRequest: String): String
    fun triangulateWiFiJson(jsonRequest: String): String
    fun analyzeIpIdJson(jsonRequest: String): String
    fun analyzeTunnelJson(jsonRequest: String): String
    fun calculateHaversineDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double
}
