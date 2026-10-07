package com.dodecaneser.rastreador.ui.theme

import android.app.Activity
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

private val TacticalDarkColorScheme = darkColorScheme(
    primary = CyberColors.CyberCyan,
    onPrimary = CyberColors.VoidBlack,
    primaryContainer = CyberColors.SurfaceElevated,
    onPrimaryContainer = CyberColors.CyberCyan,

    secondary = CyberColors.MatrixGreen,
    onSecondary = CyberColors.VoidBlack,
    secondaryContainer = CyberColors.SurfaceDark,
    onSecondaryContainer = CyberColors.MatrixGreen,

    tertiary = CyberColors.WarningAmber,
    onTertiary = CyberColors.VoidBlack,
    tertiaryContainer = CyberColors.SurfaceElevated,
    onTertiaryContainer = CyberColors.WarningAmber,

    background = CyberColors.VoidBlack,
    onBackground = CyberColors.TextPrimary,

    surface = CyberColors.SurfaceDark,
    onSurface = CyberColors.TextPrimary,
    surfaceVariant = CyberColors.SurfaceElevated,
    onSurfaceVariant = CyberColors.TextSecondary,

    error = CyberColors.HazardPink,
    onError = CyberColors.VoidBlack,
    errorContainer = CyberColors.SurfaceDark,
    onErrorContainer = CyberColors.HazardPink,

    outline = CyberColors.SurfaceBorder,
    outlineVariant = CyberColors.SurfaceBorder
)

/**
 * RastreadorTheme: Enforces strict tactical dark mode aesthetics.
 * Syncs system bars and insets with VoidBlack.
 */
@Composable
fun RastreadorTheme(
    content: @Composable () -> Unit
) {
    val colorScheme = TacticalDarkColorScheme
    val view = LocalView.current

    if (!view.isInEditMode) {
        @Suppress("DEPRECATION")
        SideEffect {
            val window = (view.context as? Activity)?.window
            if (window != null) {
                window.statusBarColor = CyberColors.VoidBlack.toArgb()
                window.navigationBarColor = CyberColors.VoidBlack.toArgb()
                val insetsController = WindowCompat.getInsetsController(window, view)
                insetsController.isAppearanceLightStatusBars = false
                insetsController.isAppearanceLightNavigationBars = false
            }
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = CyberTypography,
        content = content
    )
}
