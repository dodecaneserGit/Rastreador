# Original User Request

## Initial Request — 2026-10-07T13:41:05Z

Build a high-performance native Android application (Rastreador Mobile) combining the existing Go core networking engine with a modern tactical Jetpack Compose interface, real-time 802.11 Wi-Fi beacon scanning, 4G/5G cellular tower triangulation, and offline/online interactive mapping.

Working directory: /Volumes/SSD/Proyectos/rastreador-android
Integrity mode: development

## Requirements

### R1. Go Core Engine & Gomobile Native Bridge
Empaquetar y exportar la lógica del motor Go (pkg/recon, pkg/multilat, pkg/l2wifi, pkg/ipid, pkg/tunnel) en una biblioteca nativa Android (.aar) mediante gomobile bind o wrappers JNI C-shared, exponiendo llamadas asíncronas reactivas (Kotlin Coroutines / Flow) sin bloquear el hilo principal.

### R2. Native Hardware & Sensor Fusion (Wi-Fi, 4G/5G Cell ID, Wardriving)
Implementar servicios nativos en Android para:
- Captura pasiva de balizas Wi-Fi 802.11 en tiempo real mediante WifiManager (BSSID, SSID, frecuencia, RSSI).
- Triangulación y extracción de parámetros de torres celulares (Cell ID, TAC, LAC, MCC, MNC) con TelephonyManager.
- Modo Wardriving continuo en segundo plano con persistencia local en SQLite / Room Database.
- Detección de GPS / GNSS de alta precisión con cálculo de velocidad y rumbo.

### R3. Tactical Cyber-HUD & Interactive Touch Mapping (Jetpack Compose)
Diseñar e implementar una interfaz táctil de vanguardia en Jetpack Compose (Material 3 Dark Cyber-HUD):
- Visor de mapas táctil (OsmDroid / MapLibre) con soporte para capas Satélite, Callejero y Topográfico de Esri/Carto, marcadores dinámicos y círculos de restricción CBG.
- Radar de retardo RTT con animación de pulsos de red en tiempo real y gráficos de velocidad IP-ID.
- Exportación directa de informes forenses en PDF y JSON estructurado.

### R4. Configuración Gradle, CI/CD y Suite de Pruebas
Configurar el proyecto Gradle con Kotlin DSL (build.gradle.kts), soporte para Android 14+ (API 34/35), gestión de permisos de sistema (ACCESS_FINE_LOCATION, ACCESS_WIFI_STATE, NEARBY_WIFI_DEVICES), ProGuard/R8 y suites de pruebas unitarias e instrumentadas.

## Acceptance Criteria

### Core Engine & Interoperability
- [ ] La biblioteca Go compila limpiamente y se enlaza con el código Kotlin en la arquitectura ARM64 / x86_64.
- [ ] Las consultas BGP, sondeos RTT CBG, reconocimiento de Centrales Telefónicas y WiGLE se ejecutan de forma asíncrona mediante Coroutines.

### Hardware & Sensors
- [ ] WifiManager escanea balizas Wi-Fi del entorno y las pasa a la lista de BSSIDs para micro-triangulación.
- [ ] TelephonyManager extrae los identificadores celulares 4G/5G y permite geocodificarlos.
- [ ] El modo Wardriving almacena BSSIDs observados en la base de datos local SQLite con coordenadas y timestamp.

### UI & UX
- [ ] La interfaz Jetpack Compose es 100% fluida (60/120 fps) y adapta su layout a modo vertical y horizontal.
- [ ] El mapa interactivo muestra correctamente el punto estimado, los radios de confianza y las sondas RTT.
- [ ] El generador de reportes exporta el archivo PDF con el mapa y las tablas de telemetría forense.

### Build & Testability
- [ ] ./gradlew assembleDebug compila con éxito un APK instalable sin errores ni advertencias de lint críticas.
- [ ] La suite de pruebas unitarias (./gradlew testDebugUnitTest) valida la lógica de red, cálculos de distancia Haversine y serialización de modelos de datos.
