package com.dodecaneser.rastreador.core

import android.util.Log

/**
 * Native bridge driver using JNI C-shared library (`librastreador.so`).
 * Direct Cgo/JNI export calls to the Go networking core engine.
 */
class JniNativeBridgeDriver : NativeBridgeDriver {

    private val isLoaded: Boolean

    init {
        var loaded = false
        try {
            System.loadLibrary("rastreador")
            // Probe JNI symbol binding to verify functions are genuinely exported
            nativeCalculateHaversine(0.0, 0.0, 0.0, 0.0)
            loaded = true
        } catch (e: UnsatisfiedLinkError) {
            logWarning("librastreador not found or missing JNI symbols: ${e.message}")
            loaded = false
        } catch (e: Throwable) {
            logWarning("Failed to load librastreador: ${e.message}")
            loaded = false
        }
        isLoaded = loaded
    }

    override fun isAvailable(): Boolean = isLoaded

    override fun queryBgpAsnJson(jsonRequest: String): String {
        checkAvailable()
        return nativeQueryBgpAsn(jsonRequest)
    }

    override fun performMultilaterationJson(jsonRequest: String): String {
        checkAvailable()
        return nativePerformMultilateration(jsonRequest)
    }

    override fun triangulateWiFiJson(jsonRequest: String): String {
        checkAvailable()
        return nativeTriangulateWiFi(jsonRequest)
    }

    override fun analyzeIpIdJson(jsonRequest: String): String {
        checkAvailable()
        return nativeAnalyzeIpId(jsonRequest)
    }

    override fun analyzeTunnelJson(jsonRequest: String): String {
        checkAvailable()
        return nativeAnalyzeTunnel(jsonRequest)
    }

    override fun calculateHaversineDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
        checkAvailable()
        return nativeCalculateHaversine(lat1, lon1, lat2, lon2)
    }

    fun invokeJson(action: String, payloadJson: String): String {
        checkAvailable()
        return nativeInvokeJson(action, payloadJson)
    }

    private fun checkAvailable() {
        if (!isLoaded) {
            throw IllegalStateException("JniNativeBridgeDriver is not available in current environment")
        }
    }

    private fun logWarning(msg: String) {
        try {
            Log.w("JniNativeBridgeDriver", msg)
        } catch (_: Throwable) {
            System.err.println("JniNativeBridgeDriver: $msg")
        }
    }

    // Native JNI methods matching Go cgo exports in cmd/librastreador:
    // Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_*
    private external fun nativeQueryBgpAsn(jsonArgs: String): String
    private external fun nativePerformMultilateration(jsonArgs: String): String
    private external fun nativeTriangulateWiFi(jsonArgs: String): String
    private external fun nativeAnalyzeIpId(jsonArgs: String): String
    private external fun nativeAnalyzeTunnel(jsonArgs: String): String
    private external fun nativeCalculateHaversine(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double
    private external fun nativeInvokeJson(action: String, payloadJSON: String): String
}
