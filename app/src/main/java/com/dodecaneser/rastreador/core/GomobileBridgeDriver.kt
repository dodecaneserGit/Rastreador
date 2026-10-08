package com.dodecaneser.rastreador.core

import android.util.Log
import java.lang.reflect.Method

/**
 * Native bridge driver using Gomobile AAR bindings (`rastreador-core.aar`).
 * Dynamically links to the generated Go Mobile Java class (`Mobile.java`)
 * without requiring compile-time classpath coupling on desktop JVM test runners.
 */
class GomobileBridgeDriver : NativeBridgeDriver {

    private val mobileClass: Class<*>?
    private val queryBgpAsnMethod: Method?
    private val performMultilatMethod: Method?
    private val triangulateWiFiMethod: Method?
    private val analyzeIpIdMethod: Method?
    private val analyzeTunnelMethod: Method?
    private val calculateHaversineMethod: Method?
    private val invokeJsonMethod: Method?
    private val isLoaded: Boolean

    init {
        var cls: Class<*>? = null
        var mQueryBgp: Method? = null
        var mMultilat: Method? = null
        var mWifi: Method? = null
        var mIpId: Method? = null
        var mTunnel: Method? = null
        var mHaversine: Method? = null
        var mInvokeJson: Method? = null
        var loaded = false

        val candidateClassNames = listOf(
            "com.dodecaneser.rastreador.bridge.mobile.Mobile",
            "com.dodecaneser.rastreador.bridge.Mobile"
        )

        for (className in candidateClassNames) {
            try {
                cls = Class.forName(className)
                break
            } catch (_: ClassNotFoundException) {
                // Continue to next candidate
            } catch (_: Throwable) {
                // Ignore
            }
        }

        val targetCls = cls
        if (targetCls != null) {
            try {
                mQueryBgp = targetCls.getMethod("queryBgpAsn", String::class.java)
                mMultilat = targetCls.getMethod("performMultilateration", String::class.java)
                mWifi = targetCls.getMethod("triangulateWiFi", String::class.java)
                mIpId = targetCls.getMethod("analyzeIpId", String::class.java)
                mTunnel = targetCls.getMethod("analyzeTunnel", String::class.java)
                val doubleType = java.lang.Double.TYPE
                mHaversine = targetCls.getMethod(
                    "calculateHaversineDistance",
                    doubleType, doubleType, doubleType, doubleType
                )
                mInvokeJson = try {
                    targetCls.getMethod("invokeJson", String::class.java, String::class.java)
                } catch (_: NoSuchMethodException) {
                    null
                }

                // Probe linkage: invoke haversine with 0.0 to ensure libgojni is linked
                val testDist = mHaversine?.invoke(null, 0.0, 0.0, 0.0, 0.0) as? Double
                if (testDist != null && !testDist.isNaN()) {
                    loaded = true
                }
            } catch (e: UnsatisfiedLinkError) {
                logWarning("Gomobile class found but native libgojni.so failed to link: ${e.message}")
                loaded = false
            } catch (e: Throwable) {
                logWarning("Failed to bind Gomobile Mobile methods: ${e.message}")
                loaded = false
            }
        }

        mobileClass = cls
        queryBgpAsnMethod = mQueryBgp
        performMultilatMethod = mMultilat
        triangulateWiFiMethod = mWifi
        analyzeIpIdMethod = mIpId
        analyzeTunnelMethod = mTunnel
        calculateHaversineMethod = mHaversine
        invokeJsonMethod = mInvokeJson
        isLoaded = loaded
    }

    override fun isAvailable(): Boolean = isLoaded

    override fun queryBgpAsnJson(jsonRequest: String): String =
        invokeStringMethod(queryBgpAsnMethod, "queryBgpAsn", jsonRequest)

    override fun performMultilaterationJson(jsonRequest: String): String =
        invokeStringMethod(performMultilatMethod, "performMultilateration", jsonRequest)

    override fun triangulateWiFiJson(jsonRequest: String): String =
        invokeStringMethod(triangulateWiFiMethod, "triangulateWiFi", jsonRequest)

    override fun analyzeIpIdJson(jsonRequest: String): String =
        invokeStringMethod(analyzeIpIdMethod, "analyzeIpId", jsonRequest)

    override fun analyzeTunnelJson(jsonRequest: String): String =
        invokeStringMethod(analyzeTunnelMethod, "analyzeTunnel", jsonRequest)

    override fun calculateHaversineDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
        checkAvailable()
        return try {
            calculateHaversineMethod?.invoke(null, lat1, lon1, lat2, lon2) as? Double
                ?: throw IllegalStateException("calculateHaversineDistance method not available")
        } catch (e: Exception) {
            throw e.cause ?: e
        }
    }

    private fun invokeStringMethod(method: Method?, action: String, jsonRequest: String): String {
        checkAvailable()
        return try {
            if (method != null) {
                method.invoke(null, jsonRequest) as String
            } else if (invokeJsonMethod != null) {
                invokeJsonMethod.invoke(null, action, jsonRequest) as String
            } else {
                throw IllegalStateException("Method $action not available on Gomobile Mobile class")
            }
        } catch (e: Exception) {
            throw e.cause ?: e
        }
    }

    private fun checkAvailable() {
        if (!isLoaded) {
            throw IllegalStateException("GomobileBridgeDriver is not available in current environment")
        }
    }

    private fun logWarning(msg: String) {
        try {
            Log.w("GomobileBridgeDriver", msg)
        } catch (_: Throwable) {
            System.err.println("GomobileBridgeDriver: $msg")
        }
    }
}
