package com.dodecaneser.rastreador.ui.components

import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.dodecaneser.rastreador.ui.theme.CyberColors

/**
 * tacticalHudFrame: Draws tactical corner brackets (┌ ┐ └ ┘), bounding borders,
 * and optional CRT scanlines over any Composable container.
 */
fun Modifier.tacticalHudFrame(
    borderColor: Color = CyberColors.CyberCyan,
    cornerLength: Dp = 12.dp,
    strokeWidth: Dp = 1.5.dp,
    showScanlines: Boolean = false
): Modifier = this.drawWithContent {
    drawContent() // Render child content

    val w = size.width
    val h = size.height
    val cl = cornerLength.toPx().coerceAtMost(minOf(w, h) / 2f)
    val sw = strokeWidth.toPx()

    // 1. Subtle bounding box border
    drawRect(
        color = borderColor.copy(alpha = 0.25f),
        topLeft = Offset.Zero,
        size = Size(w, h),
        style = Stroke(width = sw * 0.5f)
    )

    // 2. High-intensity tactical corner brackets (┌ ┐ └ ┘)
    // Top-Left ┌
    drawLine(borderColor, Offset(0f, 0f), Offset(cl, 0f), strokeWidth = sw)
    drawLine(borderColor, Offset(0f, 0f), Offset(0f, cl), strokeWidth = sw)

    // Top-Right ┐
    drawLine(borderColor, Offset(w, 0f), Offset(w - cl, 0f), strokeWidth = sw)
    drawLine(borderColor, Offset(w, 0f), Offset(w, cl), strokeWidth = sw)

    // Bottom-Left └
    drawLine(borderColor, Offset(0f, h), Offset(cl, h), strokeWidth = sw)
    drawLine(borderColor, Offset(0f, h), Offset(0f, h - cl), strokeWidth = sw)

    // Bottom-Right ┘
    drawLine(borderColor, Offset(w, h), Offset(w - cl, h), strokeWidth = sw)
    drawLine(borderColor, Offset(w, h), Offset(w, h - cl), strokeWidth = sw)

    // 3. Optional CRT scanlines
    if (showScanlines) {
        var y = 0f
        while (y < h) {
            drawLine(
                color = Color.Black.copy(alpha = 0.15f),
                start = Offset(0f, y),
                end = Offset(w, y),
                strokeWidth = 1f
            )
            y += 4f
        }
    }
}
