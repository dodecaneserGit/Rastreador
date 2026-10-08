package com.dodecaneser.rastreador.sensors

import android.app.Notification
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.net.wifi.ScanResult
import android.net.wifi.WifiManager
import android.os.Build
import android.os.Bundle
import android.os.IBinder
import androidx.core.app.NotificationCompat
import androidx.core.app.ServiceCompat
import com.dodecaneser.rastreador.MainActivity
import com.dodecaneser.rastreador.R
import com.dodecaneser.rastreador.RastreadorApp
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * WardrivingService: Foreground service for continuous background RF telemetry collection.
 * Conforms to Android 14+ (API 34/35) FOREGROUND_SERVICE_TYPE_LOCATION requirements.
 */
class WardrivingService : Service(), LocationListener {

    companion object {
        const val ACTION_START = "com.dodecaneser.rastreador.action.START_WARDRIVING"
        const val ACTION_STOP = "com.dodecaneser.rastreador.action.STOP_WARDRIVING"
        private const val NOTIFICATION_ID = 4201

        private val _isServiceRunning = MutableStateFlow(false)
        val isServiceRunning: StateFlow<Boolean> = _isServiceRunning.asStateFlow()

        private val _lastLocation = MutableStateFlow<Location?>(null)
        val lastLocation: StateFlow<Location?> = _lastLocation.asStateFlow()

        private val _detectedBeaconsCount = MutableStateFlow(0)
        val detectedBeaconsCount: StateFlow<Int> = _detectedBeaconsCount.asStateFlow()
    }

    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var locationManager: LocationManager? = null
    private var wifiManager: WifiManager? = null

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        locationManager = getSystemService(Context.LOCATION_SERVICE) as? LocationManager
        wifiManager = applicationContext.getSystemService(Context.WIFI_SERVICE) as? WifiManager
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                stopWardriving()
                stopSelf()
            }
            else -> {
                startWardriving()
            }
        }
        return START_STICKY
    }

    private fun startWardriving() {
        val notification = buildNotification("Escaneando balizas RF y coordenadas GNSS...")

        val foregroundServiceType = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            ServiceInfo.FOREGROUND_SERVICE_TYPE_LOCATION
        } else {
            0
        }

        ServiceCompat.startForeground(
            this,
            NOTIFICATION_ID,
            notification,
            foregroundServiceType
        )

        _isServiceRunning.value = true
        startLocationUpdates()
    }

    private fun stopWardriving() {
        _isServiceRunning.value = false
        try {
            locationManager?.removeUpdates(this)
        } catch (e: Exception) {
            // Ignored
        }
        stopForeground(STOP_FOREGROUND_REMOVE)
    }

    private fun startLocationUpdates() {
        try {
            locationManager?.let { lm ->
                if (lm.isProviderEnabled(LocationManager.GPS_PROVIDER)) {
                    lm.requestLocationUpdates(
                        LocationManager.GPS_PROVIDER,
                        2000L,
                        1.0f,
                        this
                    )
                } else if (lm.isProviderEnabled(LocationManager.NETWORK_PROVIDER)) {
                    lm.requestLocationUpdates(
                        LocationManager.NETWORK_PROVIDER,
                        3000L,
                        2.0f,
                        this
                    )
                }
            }
        } catch (e: SecurityException) {
            // Permission missing
        }
    }

    override fun onLocationChanged(location: Location) {
        _lastLocation.value = location
        // Trigger a passive Wi-Fi harvest update
        serviceScope.launch {
            try {
                @Suppress("DEPRECATION")
                val scanResults: List<ScanResult>? = wifiManager?.scanResults
                if (scanResults != null) {
                    _detectedBeaconsCount.value = scanResults.size
                }
            } catch (e: Exception) {
                // Ignore
            }
        }
    }

    @Deprecated("Deprecated in Java")
    override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) {}
    override fun onProviderEnabled(provider: String) {}
    override fun onProviderDisabled(provider: String) {}

    private fun buildNotification(statusText: String): Notification {
        val launchIntent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            launchIntent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        return NotificationCompat.Builder(this, RastreadorApp.CHANNEL_ID_WARDRIVING)
            .setSmallIcon(android.R.drawable.ic_menu_mylocation)
            .setContentTitle(getString(R.string.wardriving_service_running))
            .setContentText(statusText)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .build()
    }

    override fun onDestroy() {
        super.onDestroy()
        stopWardriving()
        serviceScope.cancel()
    }
}
