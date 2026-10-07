package com.dodecaneser.rastreador

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.dodecaneser.rastreador.ui.components.tacticalHudFrame
import com.dodecaneser.rastreador.ui.theme.CyberColors
import com.dodecaneser.rastreador.ui.theme.RastreadorTheme

/**
 * MainActivity: Main host activity for Rastreador Mobile.
 * Configured with edge-to-edge system insets and Cyber-HUD dark theming.
 */
class MainActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        setContent {
            RastreadorTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    TacticalHudBaselineScreen()
                }
            }
        }
    }
}

/**
 * TacticalHudBaselineScreen: Milestone 1 baseline UI verifying theme, typography,
 * canvas modifiers, and resource resolution.
 */
@Composable
fun TacticalHudBaselineScreen() {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(CyberColors.VoidBlack)
            .safeDrawingPadding()
            .padding(16.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .tacticalHudFrame(
                    borderColor = CyberColors.CyberCyan,
                    cornerLength = 16.dp,
                    strokeWidth = 2.dp,
                    showScanlines = true
                )
                .background(CyberColors.SurfaceDark)
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = stringResource(R.string.hud_title),
                style = MaterialTheme.typography.displayLarge,
                color = CyberColors.CyberCyan
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = stringResource(R.string.hud_subtitle),
                style = MaterialTheme.typography.bodyMedium,
                color = CyberColors.TextSecondary
            )
            Spacer(modifier = Modifier.height(24.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = stringResource(R.string.status_core_engine),
                    style = MaterialTheme.typography.labelMedium,
                    color = CyberColors.TextMuted
                )
                Text(
                    text = stringResource(R.string.status_standby),
                    style = MaterialTheme.typography.labelMedium,
                    color = CyberColors.MatrixGreen
                )
            }
            Spacer(modifier = Modifier.height(8.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "PLATFORM TARGET",
                    style = MaterialTheme.typography.labelMedium,
                    color = CyberColors.TextMuted
                )
                Text(
                    text = stringResource(R.string.status_api_version),
                    style = MaterialTheme.typography.labelMedium,
                    color = CyberColors.CyberCyan
                )
            }
        }
    }
}
