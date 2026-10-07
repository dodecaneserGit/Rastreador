package com.dodecaneser.rastreador.core

import org.junit.Assert.*
import org.junit.Test
import kotlin.math.PI

class HaversineDistanceMathTest {

    private val driver = JvmFallbackBridgeDriver()

    @Test
    fun testZeroDistanceIdenticalCoordinates() {
        val dist = driver.calculateHaversineDistance(40.4168, -3.7038, 40.4168, -3.7038)
        assertEquals(0.0, dist, 0.00001)
    }

    @Test
    fun testEquatorOneDegreeLongitude() {
        // At equator, 1 degree of longitude is approx (2 * pi * 6371) / 360 ~ 111.19 km
        val dist = driver.calculateHaversineDistance(0.0, 0.0, 0.0, 1.0)
        assertTrue("Distance should be between 111.0 and 112.0 km, got: $dist", dist in 111.0..112.0)
    }

    @Test
    fun testEquatorOneDegreeLatitude() {
        // 1 degree of latitude is also approx 111.19 km
        val dist = driver.calculateHaversineDistance(0.0, 0.0, 1.0, 0.0)
        assertTrue("Distance should be between 111.0 and 112.0 km, got: $dist", dist in 111.0..112.0)
    }

    @Test
    fun testAntipodalDistance() {
        // Distance halfway around the globe: pi * R = pi * 6371 ~ 20015.08 km
        val dist = driver.calculateHaversineDistance(0.0, 0.0, 0.0, 180.0)
        val expected = PI * 6371.0
        assertEquals(expected, dist, 1.0)
    }

    @Test
    fun testMadridToBarcelonaDistance() {
        // Madrid: (40.4168, -3.7038), Barcelona: (41.3851, 2.1734)
        val dist = driver.calculateHaversineDistance(40.4168, -3.7038, 41.3851, 2.1734)
        assertTrue("Distance between Madrid and Barcelona should be ~505 km, got: $dist", dist in 500.0..510.0)
    }

    @Test
    fun testCrossHemisphereDistanceLondonToSydney() {
        // London: (51.5074, -0.1278), Sydney: (-33.8688, 151.2093)
        val dist = driver.calculateHaversineDistance(51.5074, -0.1278, -33.8688, 151.2093)
        assertTrue("Distance between London and Sydney should be ~16990 km, got: $dist", dist in 16900.0..17100.0)
    }

    @Test
    fun testSymmetry() {
        val d1 = driver.calculateHaversineDistance(40.4168, -3.7038, 41.3851, 2.1734)
        val d2 = driver.calculateHaversineDistance(41.3851, 2.1734, 40.4168, -3.7038)
        assertEquals(d1, d2, 0.000001)
    }
}
