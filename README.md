# 📱 Rastreador Mobile (Android)

**Aplicación Móvil Nativa de Geolocalización Activa, Multilateración IP, Reconocimiento BGP y Micro-Localización Wi-Fi/Celular**

> **Versión**: `1.0.0-rc1`  
> **Target SDK**: Android 15 (API 35) / Min SDK: Android 8.0 (API 26)  
> **Stack**: Kotlin 2.1 + Jetpack Compose (Material 3 Cyber-HUD) + OsmDroid + Room SQLite + Go Core Engine

---

## ⚡ Características Principales

1. **Reconocimiento y Multilateración IP en Tiempo Real**:
   - Consultas BGP/ASN directas y resolución de **Centrales Telefónicas FTTH (BAP/RIMA)**.
   - Sondas de retardo RTT CBG y radar animado de pulsos de red en tiempo real.
   - Huella digital de hardware (**RFC 6864** / **RFC 1323**) y detección de túneles VPN (L4 vs L7).

2. **Sensores Nativos y Micro-Localización Capa 2**:
   - **Escáner Wi-Fi en Vivo (`WifiManager`)**: Captura de balizas 802.11 del aire con potencia RSSI y frecuencia.
   - **Telemetría Celular 4G/5G (`TelephonyManager`)**: Extracción de Cell ID, TAC, LAC, MCC y MNC.
   - **Modo Wardriving Continuo**: Servicio en primer plano (`ForegroundService`) con guardado automático en base de datos SQLite (Room).
   - **Filtro de Consistencia Geodésica**: Validación cruzada con la API de **WiGLE.net** para evitar falsos positivos.

3. **Interfaz Táctil Cyber-HUD & Mapas**:
   - Tema oscuro táctico en **Jetpack Compose (Material 3)**.
   - Visor de mapas interactivo (**OsmDroid**) con soporte para capas Satelitales (Esri World Imagery), Callejero y Topográfico.
   - Exportación forense de informes en **PDF nativo** y **JSON estructurado**.

---

## 🚀 Compilación y Ejecución

El proyecto utiliza **Gradle 8.11.1** con Kotlin DSL:

```bash
# 1. Compilar y ejecutar pruebas unitarias E2E
./gradlew testDebugUnitTest

# 2. Compilar el APK de depuración
./gradlew assembleDebug

# 3. Instalar en tu dispositivo Android mediante ADB:
adb install release/rastreador-mobile-debug.apk
```

---

## 📁 Estructura del Proyecto

```text
rastreador-android/
├── app/
│   ├── build.gradle.kts           # Configuración de compilación, NDK y dependencias
│   ├── proguard-rules.pro         # Reglas de optimización R8/ProGuard
│   └── src/
│       ├── main/
│       │   ├── AndroidManifest.xml # Permisos de ubicación precisa y servicios
│       │   ├── java/com/dodecaneser/rastreador/
│       │   │   ├── MainActivity.kt           # Host Activity Edge-to-Edge
│       │   │   ├── RastreadorApp.kt          # Bootstrap de OsmDroid y Canales de Notificación
│       │   │   ├── sensors/
│       │   │   │   └── WardrivingService.kt  # Servicio de captura RF en background
│       │   │   └── ui/
│       │   │       ├── components/           # Modificadores Cyber-HUD y Canvas Radar
│       │   │       └── theme/                # Paleta táctica, tipografías y tema
│       │   └── res/                          # Iconos, temas XML y configuración de red
│       └── test/                             # Suite de 198 pruebas de aceptación E2E
├── gradle/libs.versions.toml       # Catálogo de versiones centralizado
├── release/
│   └── rastreador-mobile-debug.apk # APK compilado listo para instalar (10 MB)
└── build.gradle.kts
```

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT.
