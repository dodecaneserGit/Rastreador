package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*
import kotlinx.coroutines.*
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.io.IOException

@Serializable
private data class BgpAsnRequestDTO(
    @SerialName("ip") val ip: String,
    @SerialName("timeout_sec") val timeoutSec: Int
)

@Serializable
private data class MultilatRequestDTO(
    @SerialName("ip") val ip: String,
    @SerialName("ports") val ports: List<Int> = emptyList(),
    @SerialName("samples") val samples: Int = 3,
    @SerialName("ripe_key") val ripeKey: String = "",
    @SerialName("ripe_probes") val ripeProbes: Int = 10,
    @SerialName("timeout_sec") val timeoutSec: Int = 10
)

@Serializable
private data class TriangulateWiFiRequestDTO(
    @SerialName("beacons") val beacons: List<WiFiBeaconScan>,
    @SerialName("wigle_key") val wigleKey: String = "",
    @SerialName("timeout_sec") val timeoutSec: Int = 10
)

@Serializable
private data class AnalyzeIpIdRequestDTO(
    @SerialName("ip") val ip: String,
    @SerialName("ports") val ports: List<Int> = emptyList(),
    @SerialName("samples") val samples: Int = 10,
    @SerialName("timeout_sec") val timeoutSec: Int = 5
)

@Serializable
private data class AnalyzeTunnelRequestDTO(
    @SerialName("ip") val ip: String,
    @SerialName("l7_url") val l7Url: String = "",
    @SerialName("timeout_sec") val timeoutSec: Int = 5
)

class GoBridgeImpl(
    private val driver: NativeBridgeDriver = resolveDefaultDriver(),
    private val ioDispatcher: CoroutineDispatcher = Dispatchers.IO,
    private val json: Json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        coerceInputValues = true
        encodeDefaults = true
    }
) : RastreadorCoreEngine {

    companion object {
        fun resolveDefaultDriver(): NativeBridgeDriver {
            val jni = JniNativeBridgeDriver()
            return if (jni.isAvailable()) jni else JvmFallbackBridgeDriver()
        }
    }

    override suspend fun queryBgpAsn(ip: String, timeoutSec: Int): Result<BgpAsnResult> =
        executeBridgeCall(timeoutSec) {
            val reqJson = json.encodeToString(BgpAsnRequestDTO(ip = ip, timeoutSec = timeoutSec))
            val rawResp = driver.queryBgpAsnJson(reqJson)
            unwrapEnvelope<BgpAsnResult>(rawResp)
        }

    override suspend fun performMultilateration(
        ip: String,
        ports: List<Int>,
        samples: Int,
        ripeKey: String,
        ripeProbes: Int,
        timeoutSec: Int
    ): Result<MultilaterationResult> = executeBridgeCall(timeoutSec) {
        val reqJson = json.encodeToString(
            MultilatRequestDTO(
                ip = ip,
                ports = ports,
                samples = samples,
                ripeKey = ripeKey,
                ripeProbes = ripeProbes,
                timeoutSec = timeoutSec
            )
        )
        val rawResp = driver.performMultilaterationJson(reqJson)
        unwrapEnvelope<MultilaterationResult>(rawResp)
    }

    override suspend fun triangulateWiFi(
        beacons: List<WiFiBeaconScan>,
        wigleKey: String,
        timeoutSec: Int
    ): Result<TriangulationResult> = executeBridgeCall(timeoutSec) {
        val reqJson = json.encodeToString(
            TriangulateWiFiRequestDTO(
                beacons = beacons,
                wigleKey = wigleKey,
                timeoutSec = timeoutSec
            )
        )
        val rawResp = driver.triangulateWiFiJson(reqJson)
        unwrapEnvelope<TriangulationResult>(rawResp)
    }

    override suspend fun analyzeIpId(
        ip: String,
        ports: List<Int>,
        samples: Int,
        timeoutSec: Int
    ): Result<IpIdResult> = executeBridgeCall(timeoutSec) {
        val reqJson = json.encodeToString(
            AnalyzeIpIdRequestDTO(
                ip = ip,
                ports = ports,
                samples = samples,
                timeoutSec = timeoutSec
            )
        )
        val rawResp = driver.analyzeIpIdJson(reqJson)
        unwrapEnvelope<IpIdResult>(rawResp)
    }

    override suspend fun analyzeTunnel(
        ip: String,
        l7Url: String,
        timeoutSec: Int
    ): Result<TunnelResult> = executeBridgeCall(timeoutSec) {
        val reqJson = json.encodeToString(
            AnalyzeTunnelRequestDTO(
                ip = ip,
                l7Url = l7Url,
                timeoutSec = timeoutSec
            )
        )
        val rawResp = driver.analyzeTunnelJson(reqJson)
        unwrapEnvelope<TunnelResult>(rawResp)
    }

    override fun calculateHaversineDistance(
        lat1: Double, lon1: Double,
        lat2: Double, lon2: Double
    ): Double = driver.calculateHaversineDistance(lat1, lon1, lat2, lon2)

    /**
     * Executes bridge operation on Dispatchers.IO with strict timeout bounding,
     * re-throwing CancellationException for cooperative cancellation,
     * and wrapping other errors in Result.failure.
     */
    private suspend fun <T> executeBridgeCall(
        timeoutSec: Int,
        block: suspend () -> T
    ): Result<T> = withContext(ioDispatcher) {
        if (timeoutSec <= 0) {
            return@withContext Result.failure(
                IllegalArgumentException("Timeout must be > 0 seconds (got $timeoutSec)")
            )
        }

        try {
            withTimeout(timeoutSec * 1000L) {
                currentCoroutineContext().ensureActive()
                val result = block()
                currentCoroutineContext().ensureActive()
                Result.success(result)
            }
        } catch (e: TimeoutCancellationException) {
            Result.failure(e)
        } catch (e: CancellationException) {
            throw e // Cooperatively re-throw CancellationException!
        } catch (e: Throwable) {
            Result.failure(e)
        }
    }

    /**
     * Safely parses either an envelope BridgeEnvelope<T> or a raw JSON string of T.
     */
    private inline fun <reified T> unwrapEnvelope(rawJson: String): T {
        return try {
            val envelope = json.decodeFromString<BridgeEnvelope<T>>(rawJson)
            if (!envelope.success) {
                throw IOException(envelope.error ?: "Bridge returned failure without message")
            }
            envelope.data ?: throw IOException("Bridge returned success with null data payload")
        } catch (e: SerializationException) {
            // Direct object fallback (if Go bridge returned raw object directly)
            json.decodeFromString<T>(rawJson)
        }
    }
}
