package com.dodecaneser.rastreador.ui.screens

import android.content.Context
import android.content.Intent
import android.net.wifi.ScanResult
import android.net.wifi.WifiManager
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import com.dodecaneser.rastreador.sensors.WardrivingService
import com.dodecaneser.rastreador.ui.components.tacticalHudFrame
import com.dodecaneser.rastreador.ui.theme.CyberColors

data class UiWifiBeacon(
    val ssid: String,
    val bssid: String,
    val rssi: Int,
    val frequency: Int,
    val capabilities: String
)

@Composable
fun WardrivingScreen(
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    val isServiceRunning by WardrivingService.isServiceRunning.collectAsState()
    val lastLocation by WardrivingService.lastLocation.collectAsState()

    var wifiBeacons by remember {
        mutableStateOf(
            listOf(
                UiWifiBeacon("MiFibra-42A0", "f4:69:42:6a:ae:a0", -54, 5180, "[WPA2-PSK-CCMP]"),
                UiWifiBeacon("Vodafone_Fast_5G", "8c:ea:48:99:a1:b2", -62, 5240, "[WPA3-SAE]"),
                UiWifiBeacon("Orange-Guest", "00:1a:2b:3c:4d:5e", -68, 2412, "[WPA2-PSK]"),
                UiWifiBeacon("SmartHub_Corp", "2c:30:33:11:88:99", -74, 5745, "[WPA2-EAP]"),
                UiWifiBeacon("DIRECT-Roku-882", "50:c7:bf:12:34:56", -81, 2437, "[WPA2-PSK]")
            )
        )
    }

    // Refresh Wi-Fi scan results from system
    fun refreshLiveScans() {
        try {
            val wm = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as? WifiManager
            @Suppress("DEPRECATION")
            val results: List<ScanResult>? = wm?.scanResults
            if (!results.isNullOrEmpty()) {
                wifiBeacons = results.map { sr ->
                    UiWifiBeacon(
                        ssid = if (sr.SSID.isNullOrBlank()) "<OCULTA>" else sr.SSID,
                        bssid = sr.BSSID ?: "00:00:00:00:00:00",
                        rssi = sr.level,
                        frequency = sr.frequency,
                        capabilities = sr.capabilities ?: ""
                    )
                }.sortedByDescending { it.rssi }
            }
        } catch (e: SecurityException) {
            // Permission restricted
        }
    }

    DisposableEffect(Unit) {
        refreshLiveScans()
        onDispose { }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(CyberColors.VoidBlack)
            .padding(12.dp)
    ) {
        // --- Header Status ---
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "ESCÁNER RF & WARDRIVING",
                style = MaterialTheme.typography.displayMedium,
                color = CyberColors.CyberCyan
            )
            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(4.dp))
                    .background(if (isServiceRunning) CyberColors.MatrixGreen else CyberColors.SurfaceElevated)
                    .padding(horizontal = 6.dp, vertical = 2.dp)
            ) {
                Text(
                    text = if (isServiceRunning) "REC ON (2DO PLANO)" else "STANDBY",
                    style = MaterialTheme.typography.labelSmall.copy(
                        color = if (isServiceRunning) Color.Black else CyberColors.TextMuted
                    )
                )
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        // --- GPS & Sensor HUD Card ---
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .tacticalHudFrame(
                    borderColor = CyberColors.MatrixGreen,
                    cornerLength = 10.dp,
                    strokeWidth = 1.5.dp,
                    showScanlines = false
                )
                .background(CyberColors.SurfaceDark)
                .padding(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "GPS LOCK: ${if (lastLocation != null) "FIJADO (3D)" else "LOCALIZANDO..."}",
                    style = MaterialTheme.typography.labelMedium,
                    color = if (lastLocation != null) CyberColors.MatrixGreen else CyberColors.WarningAmber
                )
                Text(
                    text = "PRECISIÓN: ±${lastLocation?.accuracy?.let { String.format("%.1fm", it) } ?: "2.5m"}",
                    style = MaterialTheme.typography.labelMedium,
                    color = CyberColors.CyberCyan
                )
            }
            Spacer(modifier = Modifier.height(4.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "VELOCIDAD: ${lastLocation?.speed?.let { String.format("%.1f km/h", it * 3.6f) } ?: "0.0 km/h"}",
                    style = MaterialTheme.typography.labelSmall,
                    color = CyberColors.TextSecondary
                )
                Text(
                    text = "RUMBO: ${lastLocation?.bearing?.let { String.format("%.0f°", it) } ?: "045° NE"}",
                    style = MaterialTheme.typography.labelSmall,
                    color = CyberColors.TextSecondary
                )
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        // --- Wardriving Control Actions ---
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Button(
                onClick = {
                    val intent = Intent(context, WardrivingService::class.java).apply {
                        action = if (isServiceRunning) WardrivingService.ACTION_STOP else WardrivingService.ACTION_START
                    }
                    if (isServiceRunning) {
                        context.startService(intent)
                    } else {
                        ContextCompat.startForegroundService(context, intent)
                    }
                },
                modifier = Modifier.weight(1f),
                colors = ButtonDefaults.buttonColors(
                    containerColor = if (isServiceRunning) CyberColors.AlertCrimson else CyberColors.MatrixGreen,
                    contentColor = Color.Black
                ),
                shape = RoundedCornerShape(4.dp)
            ) {
                Text(
                    text = if (isServiceRunning) "DETENER WARDRIVING" else "INICIAR WARDRIVING",
                    style = MaterialTheme.typography.labelMedium
                )
            }

            Button(
                onClick = { refreshLiveScans() },
                colors = ButtonDefaults.buttonColors(
                    containerColor = CyberColors.SurfaceElevated,
                    contentColor = CyberColors.CyberCyan
                ),
                shape = RoundedCornerShape(4.dp),
                modifier = Modifier.border(1.dp, CyberColors.CyberCyan, RoundedCornerShape(4.dp))
            ) {
                Text("REFRESCAR", style = MaterialTheme.typography.labelSmall)
            }
        }

        Spacer(modifier = Modifier.height(12.dp))

        // --- Detected Beacons List ---
        Text(
            text = "BALIZAS 802.11 DETECTADAS (${wifiBeacons.size})",
            style = MaterialTheme.typography.labelMedium,
            color = CyberColors.TextSecondary
        )
        Spacer(modifier = Modifier.height(6.dp))

        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            items(wifiBeacons) { beacon ->
                WifiBeaconCard(beacon)
            }
        }
    }
}

@Composable
fun WifiBeaconCard(beacon: UiWifiBeacon) {
    val signalColor = when {
        beacon.rssi >= -60 -> CyberColors.MatrixGreen
        beacon.rssi >= -75 -> CyberColors.WarningAmber
        else -> CyberColors.AlertCrimson
    }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(4.dp))
            .background(CyberColors.SurfaceDark)
            .border(1.dp, CyberColors.SurfaceBorder, RoundedCornerShape(4.dp))
            .padding(10.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = beacon.ssid,
                style = MaterialTheme.typography.bodyLarge,
                color = Color.White
            )
            Text(
                text = "${beacon.rssi} dBm",
                style = MaterialTheme.typography.labelMedium,
                color = signalColor
            )
        }
        Spacer(modifier = Modifier.height(4.dp))
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(
                text = "BSSID: ${beacon.bssid}",
                style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                color = CyberColors.CyberCyan
            )
            Text(
                text = "${beacon.frequency} MHz",
                style = MaterialTheme.typography.labelSmall.copy(fontSize = 10.sp),
                color = CyberColors.TextMuted
            )
        }
    }
}
