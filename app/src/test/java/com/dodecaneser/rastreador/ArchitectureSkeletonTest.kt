package com.dodecaneser.rastreador

import androidx.compose.ui.text.font.FontFamily
import com.dodecaneser.rastreador.sensors.WardrivingService
import com.dodecaneser.rastreador.ui.theme.CyberColors
import com.dodecaneser.rastreador.ui.theme.CyberTypography
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import kotlin.math.abs
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.sin
import kotlin.math.sqrt

/**
 * ArchitectureSkeletonTest: Milestone 1 baseline unit verification test.
 * Validates core constants, theme color integrity, typography hierarchies,
 * service intents, and mathematical geo-distance foundations.
 */
class ArchitectureSkeletonTest {

    @Test
    fun testNotificationChannelIds() {
        assertEquals("wardriving_service_channel", RastreadorApp.CHANNEL_ID_WARDRIVING)
        assertEquals("tactical_alerts_channel", RastreadorApp.CHANNEL_ID_ALERTS)
    }

    @Test
    fun testCyberColorsIntegrity() {
        assertNotNull(CyberColors.VoidBlack)
        assertNotNull(CyberColors.SurfaceDark)
        assertNotNull(CyberColors.SurfaceElevated)
        assertNotNull(CyberColors.SurfaceBorder)
        assertNotNull(CyberColors.SurfaceGrid)
        assertNotNull(CyberColors.MatrixGreen)
        assertNotNull(CyberColors.CyberCyan)
        assertNotNull(CyberColors.WarningAmber)
        assertNotNull(CyberColors.HazardPink)
        assertNotNull(CyberColors.AlertCrimson)
        assertNotNull(CyberColors.TextPrimary)
        assertNotNull(CyberColors.TextSecondary)
        assertNotNull(CyberColors.TextMuted)
        assertNotNull(CyberColors.TextGreenGlow)
    }

    @Test
    fun testCyberTypographyHierarchy() {
        assertEquals(FontFamily.Monospace, CyberTypography.displayLarge.fontFamily)
        assertEquals(FontFamily.Monospace, CyberTypography.titleMedium.fontFamily)
        assertEquals(FontFamily.Monospace, CyberTypography.bodyMedium.fontFamily)
        assertEquals(FontFamily.Monospace, CyberTypography.labelSmall.fontFamily)

        // Verify hierarchy: displayLarge > displayMedium > titleLarge > bodyLarge > bodySmall
        assertTrue(CyberTypography.displayLarge.fontSize.value > CyberTypography.displayMedium.fontSize.value)
        assertTrue(CyberTypography.displayMedium.fontSize.value > CyberTypography.titleLarge.fontSize.value)
        assertTrue(CyberTypography.titleLarge.fontSize.value > CyberTypography.bodyLarge.fontSize.value)
        assertTrue(CyberTypography.bodyLarge.fontSize.value > CyberTypography.bodySmall.fontSize.value)
    }

    @Test
    fun testWardrivingServiceIntentActions() {
        assertEquals("com.dodecaneser.rastreador.action.START_WARDRIVING", WardrivingService.ACTION_START)
        assertEquals("com.dodecaneser.rastreador.action.STOP_WARDRIVING", WardrivingService.ACTION_STOP)
    }

    @Test
    fun testHaversineBaselineCalculation() {
        val distanceKm = calculateHaversine(0.0, 0.0, 0.0, 1.0)
        // 1 degree at the equator is approx 111.19 - 111.32 km
        assertTrue("Distance $distanceKm km out of range", distanceKm in 111.0..112.0)
    }

    @Test
    fun testHaversineZeroDistance() {
        val distanceKm = calculateHaversine(40.4168, -3.7038, 40.4168, -3.7038)
        assertEquals(0.0, distanceKm, 0.0001)
    }

    @Test
    fun testHaversineAntipodalDistance() {
        // Distance between (0, 0) and (0, 180) should be half the circumference: pi * R ~ 20015 km
        val distanceKm = calculateHaversine(0.0, 0.0, 0.0, 180.0)
        val expected = Math.PI * 6371.0
        assertTrue("Antipodal distance $distanceKm not near $expected", abs(distanceKm - expected) < 1.0)
    }

    private fun calculateHaversine(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
        val p1 = Math.toRadians(lat1)
        val p2 = Math.toRadians(lat2)
        val dLat = Math.toRadians(lat2 - lat1)
        val dLon = Math.toRadians(lon2 - lon1)

        val a = sin(dLat / 2) * sin(dLat / 2) +
                cos(p1) * cos(p2) *
                sin(dLon / 2) * sin(dLon / 2)
        val c = 2 * atan2(sqrt(a), sqrt(1 - a))
        return 6371.0 * c
    }
}
