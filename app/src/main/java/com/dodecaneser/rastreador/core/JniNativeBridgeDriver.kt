package com.dodecaneser.rastreador.core

import android.util.Log

class JniNativeBridgeDriver : NativeBridgeDriver {

    private val isLoaded: Boolean

    init {
        var loaded = false
        try {
            System.loadLibrary("rastreador")
            loaded = true
        } catch (e: UnsatisfiedLinkError) {
            try {
                Log.w("JniNativeBridgeDriver", "librastreador not found in library path: ${e.message}")
            } catch (_: Throwable) {
                System.err.println("JniNativeBridgeDriver: librastreador not found: ${e.message}")
            }
        } catch (e: Throwable) {
            try {
                Log.w("JniNativeBridgeDriver", "Failed to load librastreador: ${e.message}")
            } catch (_: Throwable) {
                System.err.println("JniNativeBridgeDriver: Failed to load librastreador: ${e.message}")
            }
        }
        isLoaded = loaded
    }

    override fun isAvailable(): Boolean = isLoaded

    override fun queryBgpAsnJson(jsonRequest: String): String = nativeQueryBgpAsn(jsonRequest)
    override fun performMultilaterationJson(jsonRequest: String): String = nativePerformMultilateration(jsonRequest)
    override fun triangulateWiFiJson(jsonRequest: String): String = nativeTriangulateWiFi(jsonRequest)
    override fun analyzeIpIdJson(jsonRequest: String): String = nativeAnalyzeIpId(jsonRequest)
    override fun analyzeTunnelJson(jsonRequest: String): String = nativeAnalyzeTunnel(jsonRequest)
    override fun calculateHaversineDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double =
        nativeCalculateHaversine(lat1, lon1, lat2, lon2)

    private external fun nativeQueryBgpAsn(jsonArgs: String): String
    private external fun nativePerformMultilateration(jsonArgs: String): String
    private external fun nativeTriangulateWiFi(jsonArgs: String): String
    private external fun nativeAnalyzeIpId(jsonArgs: String): String
    private external fun nativeAnalyzeTunnel(jsonArgs: String): String
    private external fun nativeCalculateHaversine(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double
}
