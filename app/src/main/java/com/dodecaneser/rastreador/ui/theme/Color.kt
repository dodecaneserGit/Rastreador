package com.dodecaneser.rastreador.ui.theme

import androidx.compose.ui.graphics.Color

/**
 * CyberColors: Tactical Red & High-Contrast HUD color tokens for military/recon operations.
 */
object CyberColors {
    // Void & Surface Foundations
    val VoidBlack       = Color(0xFF070A10) // Deep dark void background
    val SurfaceDark     = Color(0xFF0F1624) // Primary panel / card surface
    val SurfaceElevated = Color(0xFF182236) // Elevated controls & active tabs
    val SurfaceBorder   = Color(0xFF2B3A54) // Bounding borders
    val SurfaceGrid     = Color(0x1AFF2A4B) // Holographic gridline overlay (10% red)

    // Tactical Red Accents (Primary Theme)
    val TacticalRed     = Color(0xFFE50914) // Primary action buttons (high-visibility red)
    val TacticalCrimson = Color(0xFFFF2A4B) // Glowing borders, active indicators, titles
    val TacticalRuby    = Color(0xFFFF4D6D) // Highlights, radar sweep pulses
    val AlertCrimson    = Color(0xFFDC2626) // Errors & stop triggers

    // Secondary Auxiliary Indicators
    val MatrixGreen     = Color(0xFF10B981) // Nominal GPS Lock, live status badge
    val CyberCyan       = Color(0xFF00E5FF) // Coordinates & data highlights
    val WarningAmber    = Color(0xFFF59E0B) // Medium signal, warning alerts
    val HazardPink      = Color(0xFFFF0055) // Critical alerts

    // Monospaced High-Contrast Text Hierarchy
    val PureWhite       = Color(0xFFFFFFFF) // Maximum contrast for button labels & critical data
    val TextPrimary     = Color(0xFFFFFFFF) // High-contrast data readouts
    val TextSecondary   = Color(0xFFCBD5E1) // Metric labels & units
    val TextMuted       = Color(0xFF8896AB) // Inactive text & descriptions
    val TextGreenGlow   = Color(0xFF10B981) // Phosphor CRT terminal text
}
