package com.dodecaneser.rastreador.ui

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.core.content.ContextCompat
import com.dodecaneser.rastreador.ui.components.TacticalBottomNavigationBar
import com.dodecaneser.rastreador.ui.components.TacticalTab
import com.dodecaneser.rastreador.ui.screens.NetworkIntelScreen
import com.dodecaneser.rastreador.ui.screens.RadarMapScreen
import com.dodecaneser.rastreador.ui.screens.WardrivingScreen
import com.dodecaneser.rastreador.ui.theme.CyberColors

@Composable
fun MainTacticalApp() {
    val context = LocalContext.current
    var selectedTab by remember { mutableStateOf(TacticalTab.RADAR_MAP) }

    // Runtime permissions for Location, Wi-Fi & Notifications
    val permissionsToRequest = remember {
        mutableListOf(
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION
        ).apply {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                add(Manifest.permission.POST_NOTIFICATIONS)
                add(Manifest.permission.NEARBY_WIFI_DEVICES)
            }
        }.toTypedArray()
    }

    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestMultiplePermissions()
    ) { _ ->
        // Permissions handled
    }

    LaunchedEffect(Unit) {
        val needsRequest = permissionsToRequest.any {
            ContextCompat.checkSelfPermission(context, it) != PackageManager.PERMISSION_GRANTED
        }
        if (needsRequest) {
            permissionLauncher.launch(permissionsToRequest)
        }
    }

    Scaffold(
        bottomBar = {
            TacticalBottomNavigationBar(
                selectedTab = selectedTab,
                onTabSelected = { selectedTab = it }
            )
        },
        containerColor = CyberColors.VoidBlack
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(CyberColors.VoidBlack)
                .safeDrawingPadding()
                .padding(innerPadding)
        ) {
            when (selectedTab) {
                TacticalTab.RADAR_MAP -> RadarMapScreen()
                TacticalTab.WARDRIVING -> WardrivingScreen()
                TacticalTab.INTEL -> NetworkIntelScreen()
            }
        }
    }
}
