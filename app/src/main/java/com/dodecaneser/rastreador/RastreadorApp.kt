package com.dodecaneser.rastreador

import android.app.Application
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.os.Build
import org.osmdroid.config.Configuration
import java.io.File

/**
 * RastreadorApp: Main application class for Rastreador Mobile.
 * 
 * Bootstraps system-wide services before any Activity or Service launches:
 * 1. OsmDroid map tile cache and user-agent initialization.
 * 2. Android 8.0+ (API 26+) notification channels for background wardriving and tactical alerts.
 */
class RastreadorApp : Application() {

    companion object {
        const val CHANNEL_ID_WARDRIVING = "wardriving_service_channel"
        const val CHANNEL_ID_ALERTS = "tactical_alerts_channel"

        lateinit var instance: RastreadorApp
            private set
    }

    override fun onCreate() {
        super.onCreate()
        instance = this

        initOsmDroid()
        initNotificationChannels()
    }

    /**
     * Configures OsmDroid tile engine with private internal caching and custom user-agent.
     */
    private fun initOsmDroid() {
        val sharedPrefs = getSharedPreferences("osmdroid_prefs", Context.MODE_PRIVATE)
        val osmConfig = Configuration.getInstance()
        osmConfig.load(this, sharedPrefs)

        // Set tactical user agent header for map tile downloads
        osmConfig.userAgentValue = "RastreadorMobile/1.0 (Android; Reconnaissance Client)"

        // Anchor cache inside internal application storage to comply with Scoped Storage
        val osmBasePath = File(cacheDir, "osmdroid")
        val osmTileCache = File(osmBasePath, "tiles")
        if (!osmTileCache.exists()) {
            osmTileCache.mkdirs()
        }
        osmConfig.osmdroidBasePath = osmBasePath
        osmConfig.osmdroidTileCache = osmTileCache
    }

    /**
     * Registers system notification channels for Android 8.0+ (API 26+).
     */
    private fun initNotificationChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val notificationManager = getSystemService(NotificationManager::class.java)
                ?: return

            // 1. Channel for Wardriving Foreground Service (Continuous, silent background scanning)
            val wardrivingChannel = NotificationChannel(
                CHANNEL_ID_WARDRIVING,
                getString(R.string.notification_channel_wardriving_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.notification_channel_wardriving_desc)
                setShowBadge(false)
                enableLights(false)
                enableVibration(false)
            }

            // 2. Channel for High-Priority Tactical Alerts & Forensic Dossier Exports
            val alertsChannel = NotificationChannel(
                CHANNEL_ID_ALERTS,
                getString(R.string.notification_channel_alerts_name),
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = getString(R.string.notification_channel_alerts_desc)
                setShowBadge(true)
                enableLights(true)
                enableVibration(true)
            }

            notificationManager.createNotificationChannel(wardrivingChannel)
            notificationManager.createNotificationChannel(alertsChannel)
        }
    }
}
