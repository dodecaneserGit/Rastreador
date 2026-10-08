package com.dodecaneser.rastreador

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.dodecaneser.rastreador.ui.MainTacticalApp
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
                    MainTacticalApp()
                }
            }
        }
    }
}
