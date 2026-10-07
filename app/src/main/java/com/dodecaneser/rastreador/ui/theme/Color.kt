package com.dodecaneser.rastreador.ui.theme

import androidx.compose.ui.graphics.Color

/**
 * CyberColors: Tactical HUD color tokens for military/recon operations.
 */
object CyberColors {
    // Void & Surface Foundations
    val VoidBlack       = Color(0xFF050811) // Deepest background / absolute dark void
    val SurfaceDark     = Color(0xFF0D1322) // Primary panel / card surface
    val SurfaceElevated = Color(0xFF151E33) // Floating dialogs, elevated controls, dropdowns
    val SurfaceBorder   = Color(0xFF1E2D4A) // Subtle HUD bounding borders
    val SurfaceGrid     = Color(0x1A00E5FF) // Holographic gridline overlay (10% cyan)

    // Tactical Phosphor Accents
    val MatrixGreen     = Color(0xFF00FF66) // Nominal status, Direct IP, strong Wi-Fi (>-60dBm)
    val CyberCyan       = Color(0xFF00E5FF) // Active scanner, RTT radar sweep, local vantage pins
    val WarningAmber    = Color(0xFFFFB800) // Cloud ASNs, medium signal, hop landmarks
    val HazardPink      = Color(0xFFFF0055) // Critical alerts, VPN/Proxy, target pin
    val AlertCrimson    = Color(0xFFEF4444) // Confidence circle stroke, errors

    // Monospaced Text Hierarchy
    val TextPrimary     = Color(0xFFF0F6FC) // High-contrast data readouts & coordinates
    val TextSecondary   = Color(0xFF8B9CB6) // Metric labels & units
    val TextMuted       = Color(0xFF4A5B73) // Disabled / inactive status
    val TextGreenGlow   = Color(0xFF39FF14) // Phosphor CRT terminal text
}
