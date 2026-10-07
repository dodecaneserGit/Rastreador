package com.dodecaneser.rastreador.sensors

import android.app.Service
import android.content.Intent
import android.os.IBinder

/**
 * WardrivingService: Foreground service for continuous background RF telemetry collection.
 * Specialized with foregroundServiceType="location" under Android 14+ (API 34/35).
 */
class WardrivingService : Service() {

    companion object {
        const val ACTION_START = "com.dodecaneser.rastreador.action.START_WARDRIVING"
        const val ACTION_STOP = "com.dodecaneser.rastreador.action.STOP_WARDRIVING"
    }

    override fun onBind(intent: Intent?): IBinder? {
        return null
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        return START_STICKY
    }
}
