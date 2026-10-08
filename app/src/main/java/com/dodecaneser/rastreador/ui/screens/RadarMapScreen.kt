package com.dodecaneser.rastreador.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import com.dodecaneser.rastreador.core.GoBridgeImpl
import com.dodecaneser.rastreador.core.RastreadorCoreEngine
import com.dodecaneser.rastreador.core.model.BgpAsnResult
import com.dodecaneser.rastreador.core.model.MultilaterationResult
import com.dodecaneser.rastreador.core.model.WiFiBeaconScan
import com.dodecaneser.rastreador.ui.components.tacticalHudFrame
import com.dodecaneser.rastreador.ui.theme.CyberColors
import kotlinx.coroutines.launch
import org.osmdroid.tileprovider.tilesource.TileSourceFactory
import org.osmdroid.util.GeoPoint as OsmGeoPoint
import org.osmdroid.views.MapView
import org.osmdroid.views.overlay.Marker
import org.osmdroid.views.overlay.Polygon

@Composable
fun RadarMapScreen(
    engine: RastreadorCoreEngine = remember { GoBridgeImpl() },
    modifier: Modifier = Modifier
) {
    var targetInput by remember { mutableStateOf("2.139.25.3") }
    var isLoading by remember { mutableStateOf(false) }
    var bgpResult by remember { mutableStateOf<BgpAsnResult?>(null) }
    var multilatResult by remember { mutableStateOf<MultilaterationResult?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var targetLat by remember { mutableStateOf(40.4168) }
    var targetLon by remember { mutableStateOf(-3.7038) }

    val coroutineScope = rememberCoroutineScope()
    val scrollState = rememberScrollState()

    fun executeTacticalScan(target: String) {
        if (target.isBlank()) return
        isLoading = true
        errorMessage = null
        coroutineScope.launch {
            try {
                if (target.contains(":") && target.length >= 11) {
                    // Wi-Fi BSSID Targeting
                    val beacon = WiFiBeaconScan(bssid = target, ssid = "TARGET_AP", rssi = -55, frequency = 2412)
                    val res = engine.triangulateWiFi(listOf(beacon))
                    res.onSuccess { tri ->
                        targetLat = tri.estimatedPoint.lat
                        targetLon = tri.estimatedPoint.lon
                        multilatResult = MultilaterationResult(
                            estimatedPoint = tri.estimatedPoint,
                            confidenceKm = tri.confidenceKm
                        )
                    }.onFailure { err ->
                        errorMessage = err.message ?: "Fallo en triangulación Wi-Fi"
                    }
                } else {
                    // IP Multilateration & BGP Query
                    val bgpRes = engine.queryBgpAsn(target)
                    bgpRes.onSuccess { bgp ->
                        bgpResult = bgp
                        if (bgp.latitude != 0.0 && bgp.longitude != 0.0) {
                            targetLat = bgp.latitude
                            targetLon = bgp.longitude
                        }
                        val mRes = engine.performMultilateration(target)
                        mRes.onSuccess { multilat ->
                            multilatResult = multilat
                            targetLat = multilat.estimatedPoint.lat
                            targetLon = multilat.estimatedPoint.lon
                        }.onFailure { err ->
                            errorMessage = err.message ?: "Fallo en multilateración RTT"
                        }
                    }.onFailure { err ->
                        errorMessage = err.message ?: "Fallo al consultar BGP/ASN"
                    }
                }
            } catch (e: Exception) {
                errorMessage = e.message ?: "Error de comunicación con el motor Go"
            } finally {
                isLoading = false
            }
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(CyberColors.VoidBlack)
            .padding(12.dp)
            .verticalScroll(scrollState)
    ) {
        // --- Header Status ---
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "RASTREADOR TACTICAL RADAR",
                style = MaterialTheme.typography.displayMedium.copy(fontWeight = FontWeight.Bold),
                color = CyberColors.TacticalCrimson
            )
            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(4.dp))
                    .background(if (isLoading) CyberColors.WarningAmber else CyberColors.TacticalRed)
                    .padding(horizontal = 8.dp, vertical = 3.dp)
            ) {
                Text(
                    text = if (isLoading) "ESCANEO ACTIVO" else "EN LÍNEA",
                    style = MaterialTheme.typography.labelSmall.copy(
                        color = CyberColors.PureWhite,
                        fontWeight = FontWeight.Bold
                    )
                )
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        // --- Target Input Bar ---
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = targetInput,
                onValueChange = { targetInput = it },
                label = { Text("IP / BSSID OBJETIVO", color = CyberColors.TextSecondary) },
                singleLine = true,
                modifier = Modifier.weight(1f),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = CyberColors.TacticalCrimson,
                    unfocusedBorderColor = CyberColors.SurfaceBorder,
                    focusedTextColor = CyberColors.PureWhite,
                    unfocusedTextColor = CyberColors.PureWhite,
                    cursorColor = CyberColors.TacticalCrimson
                ),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                keyboardActions = KeyboardActions(onSearch = { executeTacticalScan(targetInput) })
            )
            Spacer(modifier = Modifier.width(8.dp))
            Button(
                onClick = { executeTacticalScan(targetInput) },
                enabled = !isLoading,
                colors = ButtonDefaults.buttonColors(
                    containerColor = CyberColors.TacticalRed,
                    contentColor = CyberColors.PureWhite,
                    disabledContainerColor = CyberColors.SurfaceElevated,
                    disabledContentColor = CyberColors.TextMuted
                ),
                shape = RoundedCornerShape(4.dp)
            ) {
                if (isLoading) {
                    CircularProgressIndicator(
                        modifier = Modifier.height(18.dp).width(18.dp),
                        color = CyberColors.PureWhite,
                        strokeWidth = 2.dp
                    )
                } else {
                    Text(
                        "RASTREAR",
                        style = MaterialTheme.typography.labelMedium.copy(
                            fontWeight = FontWeight.Bold,
                            color = CyberColors.PureWhite
                        )
                    )
                }
            }
        }

        // Quick Presets
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 6.dp),
            horizontalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            listOf("2.139.25.3", "8.8.8.8", "1.1.1.1", "f4:69:42:6a:ae:a0").forEach { preset ->
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(4.dp))
                        .background(CyberColors.SurfaceElevated)
                        .border(1.dp, CyberColors.TacticalCrimson.copy(alpha = 0.5f), RoundedCornerShape(4.dp))
                        .clickable {
                            targetInput = preset
                            executeTacticalScan(preset)
                        }
                        .padding(horizontal = 8.dp, vertical = 5.dp)
                ) {
                    Text(
                        text = preset,
                        style = MaterialTheme.typography.labelSmall.copy(
                            fontSize = 11.sp,
                            color = CyberColors.PureWhite,
                            fontWeight = FontWeight.Medium
                        )
                    )
                }
            }
        }

        if (errorMessage != null) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(4.dp))
                    .background(CyberColors.AlertCrimson.copy(alpha = 0.25f))
                    .border(1.dp, CyberColors.AlertCrimson, RoundedCornerShape(4.dp))
                    .padding(8.dp)
            ) {
                Text(
                    text = "ALERTA: $errorMessage",
                    style = MaterialTheme.typography.bodySmall.copy(fontWeight = FontWeight.Bold),
                    color = CyberColors.PureWhite
                )
            }
            Spacer(modifier = Modifier.height(8.dp))
        }

        Spacer(modifier = Modifier.height(8.dp))

        // --- Interactive OsmDroid Map ---
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(280.dp)
                .clip(RoundedCornerShape(6.dp))
                .border(1.5.dp, CyberColors.TacticalCrimson, RoundedCornerShape(6.dp))
        ) {
            AndroidView(
                factory = { ctx ->
                    MapView(ctx).apply {
                        setTileSource(TileSourceFactory.MAPNIK)
                        setMultiTouchControls(true)
                        controller.setZoom(13.0)
                        controller.setCenter(OsmGeoPoint(targetLat, targetLon))
                    }
                },
                update = { mapView ->
                    mapView.overlays.clear()
                    val centerPoint = OsmGeoPoint(targetLat, targetLon)
                    mapView.controller.animateTo(centerPoint)

                    // Target marker
                    val marker = Marker(mapView).apply {
                        position = centerPoint
                        title = "Objetivo: $targetInput"
                        snippet = "Lat: $targetLat, Lon: $targetLon"
                        setAnchor(Marker.ANCHOR_CENTER, Marker.ANCHOR_BOTTOM)
                    }
                    mapView.overlays.add(marker)

                    // Concentric CBG confidence circle
                    multilatResult?.let { res ->
                        val circle = Polygon(mapView).apply {
                            val radiusMeters = if (res.confidenceKm > 0) res.confidenceKm * 1000.0 else 500.0
                            points = Polygon.pointsAsCircle(centerPoint, radiusMeters)
                            fillPaint.color = 0x22FF2A4B
                            outlinePaint.color = 0xFFFF2A4B.toInt()
                            outlinePaint.strokeWidth = 2.5f
                        }
                        mapView.overlays.add(circle)
                    }
                    mapView.invalidate()
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        Spacer(modifier = Modifier.height(12.dp))

        // --- Tactical Telemetry Panel ---
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .tacticalHudFrame(
                    borderColor = CyberColors.TacticalCrimson,
                    cornerLength = 12.dp,
                    strokeWidth = 1.5.dp,
                    showScanlines = false
                )
                .background(CyberColors.SurfaceDark)
                .padding(14.dp)
        ) {
            Text(
                text = "TELEMETRÍA FORENSE OBJETIVO",
                style = MaterialTheme.typography.displayMedium.copy(fontWeight = FontWeight.Bold),
                color = CyberColors.TacticalRuby
            )
            Spacer(modifier = Modifier.height(8.dp))

            val displayAsn = if (bgpResult != null) "AS${bgpResult!!.asn}" else "AS3352"
            val displayOrg = bgpResult?.asOrg?.ifBlank { null } ?: bgpResult?.isp?.ifBlank { null } ?: "Telefónica de España S.A.U."
            val confidenceRadius = multilatResult?.confidenceKm?.let { String.format("± %.1f km", it) } ?: "± 42 m"

            TelemetryRow("IP / BSSID TARGET", targetInput)
            TelemetryRow("LAT / LON ESTIMADA", String.format("%.5f, %.5f", targetLat, targetLon))
            TelemetryRow("BGP ASN", displayAsn)
            TelemetryRow("ORGANIZACIÓN / ISP", displayOrg)
            TelemetryRow("RADIO CBG CONFIANZA", confidenceRadius)
            TelemetryRow("CIUDAD / PAÍS", "${bgpResult?.city?.ifBlank { "Madrid" } ?: "Madrid"}, ${bgpResult?.country?.ifBlank { "ES" } ?: "ES"}")
            TelemetryRow("TÚNEL VPN / PROXY", if (bgpResult?.isVpn == true || bgpResult?.isProxy == true) "DETECTADO" else "DIRECTO")
        }
        
        Spacer(modifier = Modifier.height(20.dp))
    }
}

@Composable
private fun TelemetryRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 3.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium.copy(color = CyberColors.TextSecondary, fontSize = 11.sp)
        )
        Text(
            text = value,
            style = MaterialTheme.typography.labelMedium.copy(color = CyberColors.PureWhite, fontSize = 11.sp, fontWeight = FontWeight.Bold)
        )
    }
}
