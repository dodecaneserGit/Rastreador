package com.dodecaneser.rastreador.core

import com.dodecaneser.rastreador.core.model.*

/**
 * High-performance tactical network reconnaissance engine interface.
 * Exposes non-blocking coroutines and reactive Result<T> containers.
 */
interface RastreadorCoreEngine {

    suspend fun queryBgpAsn(
        ip: String,
        timeoutSec: Int = 5
    ): Result<BgpAsnResult>

    suspend fun performMultilateration(
        ip: String,
        ports: List<Int> = listOf(80, 443, 53, 22),
        samples: Int = 3,
        ripeKey: String = "",
        ripeProbes: Int = 10,
        timeoutSec: Int = 10
    ): Result<MultilaterationResult>

    suspend fun triangulateWiFi(
        beacons: List<WiFiBeaconScan>,
        wigleKey: String = "",
        timeoutSec: Int = 10
    ): Result<TriangulationResult>

    suspend fun analyzeIpId(
        ip: String,
        ports: List<Int> = listOf(80, 443, 22),
        samples: Int = 10,
        timeoutSec: Int = 5
    ): Result<IpIdResult>

    suspend fun analyzeTunnel(
        ip: String,
        l7Url: String = "",
        timeoutSec: Int = 5
    ): Result<TunnelResult>

    fun calculateHaversineDistance(
        lat1: Double, lon1: Double,
        lat2: Double, lon2: Double
    ): Double
}
