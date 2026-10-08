# 📖 MANUAL DE USO Y OPERACIÓN — RASTREADOR MOBILE (ANDROID)

**Manual Técnico y Operativo Completo para la Suite Móvil de Geolocalización Activa, Multilateración IP, Reconocimiento BGP y Wardriving RF**

---

## 📑 ÍNDICE DE CONTENIDOS

1. [Introducción y Arquitectura General](#1-introducción-y-arquitectura-general)
2. [Requisitos e Instalación del APK](#2-requisitos-e-instalación-del-apk)
3. [Permisos del Sistema Android y Primer Inicio](#3-permisos-del-sistema-android-y-primer-inicio)
4. [Navegación e Interfaz Cyber-HUD](#4-navegación-e-interfaz-cyber-hud)
5. [Módulo 1: Radar / Mapa Táctico & Multilateración (GEO)](#5-módulo-1-radar--mapa-táctico--multilateración-geo)
6. [Módulo 2: Escáner RF, Telemetría y Wardriving (802.11)](#6-módulo-2-escáner-rf-telemetría-y-wardriving-80211)
7. [Módulo 3: Inteligencia de Red BGP, IP-ID y Túneles VPN (BGP/IP-ID)](#7-módulo-3-inteligencia-de-red-bgp-ip-id-y-túneles-vpn-bgpip-id)
8. [Resolución de Problemas Frecuentes (Troubleshooting)](#8-resolución-de-problemas-frecuentes-troubleshooting)
9. [Especificaciones de Seguridad y Privacidad](#9-especificaciones-de-seguridad-y-privacidad)

---

## 1. Introducción y Arquitectura General

**Rastreador Mobile** es una plataforma nativa para dispositivos Android desarrollada para profesionales de ciberseguridad, analistas de inteligencia de fuentes abiertas (OSINT) y auditores de redes.

Combina dos núcleos fundamentales:
- **Motor de Red en Go (`pkg/*`):** Procesa cálculos geodésicos de alta precisión (Haversine), consultas a registros RIR/BGP, sondas de retardo RTT (*Round-Trip Time*), análisis de huella digital de hardware IP-ID (**RFC 6864**) y detección de discrepancias L4/L7 para túneles VPN.
- **Interfaz y Sensores en Kotlin/Jetpack Compose:** Ofrece una experiencia táctica de alto contraste en **Rojo Táctico** y **Void Black**, integración directa con el hardware Wi-Fi/GNSS del dispositivo y visor de mapas OpenStreetMap/OsmDroid.

---

## 2. Requisitos e Instalación del APK

### Requisitos del Dispositivo
* **Sistema Operativo:** Android 8.0 (Oreo / API 26) hasta Android 15 (Vanilla Ice Cream / API 35).
* **Hardware:**
  * Receptor GNSS / GPS integrado (recomendado para fijación de coordenadas de campo).
  * Adaptador Wi-Fi 802.11 b/g/n/ac/ax con soporte de escaneo pasivo.
  * Módulo celular 4G LTE / 5G NR (opcional, para telemetría de torres).

### Métodos de Instalación

#### Opción A: Instalación directa mediante ADB (Recomendado para auditores)
Conecta tu dispositivo Android al ordenador por cable USB con la **Depuración USB** activada y ejecuta:

```bash
adb install -r /Volumes/SSD/Proyectos/rastreador-android/release/rastreador-mobile-debug.apk
```

#### Opción B: Instalación manual en el dispositivo
1. Copia el archivo `rastreador-mobile-debug.apk` a la memoria interna de tu móvil o descárgalo directamente.
2. Abre tu explorador de archivos en Android y pulsa sobre el APK.
3. Si el sistema lo solicita, autoriza la opción *"Instalar aplicaciones de fuentes desconocidas"*.
4. Pulsa en **Instalar** y luego en **Abrir**.

---

## 3. Permisos del Sistema Android y Primer Inicio

Al abrir la aplicación por primera vez, el sistema solicitará automáticamente los siguientes permisos indispensables:

1. **Ubicación Precisa (`ACCESS_FINE_LOCATION` y `ACCESS_COARSE_LOCATION`):**
   * *¿Para qué se usa?* Permite calcular las coordenadas geodésicas del dispositivo, determinar el radio de precisión métrico (±2.5m), la velocidad y el rumbo durante las auditorías de campo.
   * *Acción:* Selecciona **"Mientras la app esté en uso"** o **"Permitir siempre"**.

2. **Dispositivos Wi-Fi Cercanos (`NEARBY_WIFI_DEVICES` en Android 13+):**
   * *¿Para qué se usa?* Permite a la aplicación escuchar balizas 802.11 del entorno y capturar las direcciones MAC (BSSID) y potencias (RSSI) sin asociarse a las redes.
   * *Acción:* Selecciona **"Permitir"**.

3. **Notificaciones del Sistema (`POST_NOTIFICATIONS` en Android 13+):**
   * *¿Para qué se usa?* Mantiene activa la notificación táctica en la barra de estado cuando ejecutas el **Servicio de Wardriving en Segundo Plano**.
   * *Acción:* Selecciona **"Permitir"**.

---

## 4. Navegación e Interfaz Cyber-HUD

La aplicación está organizada en **tres módulos principales** accesibles desde la **Barra de Navegación Táctica Inferior**:

| Botón de Pestaña | Etiqueta | Función Principal |
| :--- | :---: | :--- |
| **`[ GEO ]`** | **RADAR / MAPA** | Geolocalización de IPs, micro-triangulación de BSSIDs, círculos de restricción CBG y mapa táctico. |
| **`[ 802.11 ]`** | **WARDRIVING RF** | Escaneo en vivo de redes Wi-Fi, telemetría GPS/GNSS y control del servicio de registro en segundo plano. |
| **`[ BGP/IP-ID ]`** | **INTEL RED** | Análisis BGP/ASN, auditoría de velocidad IP-ID secuencial (RFC 6864) y detección de fugas en túneles VPN. |

* **Indicador de Pestaña Activa:** La pestaña seleccionada se resalta en **Rojo Táctico (`#E50914`)** con tipografía en **Blanco Puro (`#FFFFFF`)**.

---

## 5. Módulo 1: Radar / Mapa Táctico & Multilateración (GEO)

Este módulo permite geolocalizar objetivos en tiempo real tanto a nivel de red pública (direcciones IP) como a nivel de capa 2 (direcciones MAC de routers Wi-Fi).

<p align="center">
  <img src="screenshots/hud_map.jpg" width="380" alt="Pantalla Radar Mapa Táctico">
</p>

### Pasos para realizar un rastreo:

1. **Ingreso del Objetivo:**
   * Introduce en el cuadro de texto **`IP / BSSID OBJETIVO`** una dirección IPv4 (ej: `2.139.25.3`), IPv6 o una dirección MAC de router Wi-Fi (ej: `f4:69:42:6a:ae:a0`).
   * *Atajos rápidos:* Puedes pulsar directamente sobre cualquiera de los chips de presets rápidos (`2.139.25.3`, `8.8.8.8`, `1.1.1.1`) para autocompletar e iniciar el análisis de inmediato.

2. **Ejecución:**
   * Pulsa el botón rojo **`RASTREAR`**.
   * El indicador superior cambiará a **`ESCANEO ACTIVO`** mientras el motor Go realiza las consultas DNS, BGP y el cálculo geodésico.

3. **Interpretación del Mapa Interactivo:**
   * **Marcador Rojo Pin:** Señala las coordenadas estimadas del objetivo (latitud y longitud).
   * **Círculo Concéntrico Rojo (Restricción CBG):** Representa el radio de confianza geodésico calculado. Cuanto menor sea el radio, mayor será la certeza de la ubicación física.
   * **Interacción:** Puedes hacer zoom con dos dedos (pellizco) o arrastrar el mapa para explorar calles, manzanas y puntos de interés colindantes.

4. **Lectura de la Telemetría Forense Objetivo:**
   * **`IP / BSSID TARGET`**: Identificador auditado.
   * **`LAT / LON ESTIMADA`**: Coordenadas en formato decimal con 5 decimales de precisión.
   * **`BGP ASN`**: Número de Sistema Autónomo que anuncia el prefijo (ej: `AS3352`).
   * **`ORGANIZACIÓN / ISP`**: Proveedor de servicios o entidad propietaria del bloque de red.
   * **`RADIO CBG CONFIANZA`**: Margen de incertidumbre en metros o kilómetros.
   * **`CIUDAD / PAÍS`**: Geocodificación estimada.
   * **`TÚNEL VPN / PROXY`**: Indica si el objetivo presenta indicios de anonimizador o conexión directa.

---

## 6. Módulo 2: Escáner RF, Telemetría y Wardriving (802.11)

Diseñado para auditorías inalámbricas de campo (*wardriving*), levantamiento de puntos de acceso y comprobación de cobertura de señal.

<p align="center">
  <img src="screenshots/wardriving.jpg" width="380" alt="Pantalla Escáner Wardriving">
</p>

### Funcionalidades del Módulo:

1. **Tarjeta de Fijación GPS / GNSS:**
   * **`GPS LOCK`**: Estado de sincronización satelital. Muestra `FIJADO (3D)` cuando la señal es óptima.
   * **`PRECISIÓN`**: Radio de error métrico del sensor GPS interno (ej: `±2.5m`).
   * **`VELOCIDAD`**: Velocidad de desplazamiento calculada en kilómetros por hora (`km/h`).
   * **`RUMBO`**: Orientación angular en grados y punto cardinal (`045° NE`).

2. **Modo Wardriving en Segundo Plano (`ForegroundService`):**
   * **Iniciar Captura Continua:** Pulsa el botón rojo **`INICIAR WARDRIVING`**. El sistema lanzará un servicio persistente que continuará registrando balizas Wi-Fi y coordenadas GPS incluso si bloqueas la pantalla o cambias de aplicación.
   * **Notificación Persistente:** En la cortina de notificaciones de Android aparecerá el aviso *"Tactical Wardriving Active - Escaneando balizas RF y coordenadas GNSS"*.
   * **Detener Captura:** Pulsa el botón **`DETENER WARDRIVING`** para liberar los receptores de radio y detener el registro en base de datos.

3. **Escaneo Manual y Lista de Balizas Wi-Fi:**
   * Pulsa **`REFRESCAR`** para forzar una lectura instantánea de los puntos de acceso del entorno.
   * Cada tarjeta de baliza Wi-Fi incluye:
     * **Nombre de Red (SSID):** Identificador emitido o `<OCULTA>` si tiene el broadcast desactivado.
     * **Potencia de Señal (RSSI en dBm):**
       * 🟢 **Verde (>-60 dBm):** Señal excelente y muy cercana.
       * 🟡 **Ámbar (-60 dBm a -75 dBm):** Señal media / distancia intermedia.
       * 🔴 **Rojo (<-75 dBm):** Señal débil o lejana.
     * **Dirección MAC (BSSID):** Identificador físico único del punto de acceso.
     * **Frecuencia Operativa:** Banda de 2.4 GHz (ej: `2412 MHz` / Canal 1) o 5 GHz (ej: `5180 MHz` / Canal 36).

---

## 7. Módulo 3: Inteligencia de Red BGP, IP-ID y Túneles VPN (BGP/IP-ID)

Especializado en la inspección forense no invasiva de servidores y hosts remotos para determinar su arquitectura de red y estado de actividad.

<p align="center">
  <img src="screenshots/core_analysis.jpg" width="380" alt="Pantalla Inteligencia de Red BGP e IP-ID">
</p>

### Guía de Auditoría Forense:

1. **Lanzar la Auditoría:**
   * Introduce la dirección IP a inspeccionar en el campo **`IP DE OBJETIVO`**.
   * Pulsa el botón rojo **`AUDITAR`**.

2. **Interpretación de Resultados:**

   #### 🅰️ Tarjeta 1: Enrutamiento BGP / Sistema Autónomo
   * **`SISTEMA AUTÓNOMO (ASN)`**: Número oficial de AS en los registros globales (RIPE, ARIN, LACNIC, APNIC).
   * **`ORGANIZACIÓN / AS`**: Empresa u operador que controla el enrutamiento BGP.
   * **`HOSTNAME / DNS`**: Nombre de dominio canónico obtenido por resolución inversa (rDNS).
   * **`PAÍS / CIUDAD`**: Ubicación geográfica del PoP (*Point of Presence*).

   #### 🅱️ Tarjeta 2: Velocidad IP-ID y Huella de Hardware (RFC 6864)
   * **`VELOCIDAD GLOBAL IP-ID`**: Cantidad de paquetes que el kernel del objetivo genera por segundo (ej: `24,198 IDs/s`).
   * **`LINEALIDAD Y AJUSTE R²`**: Coeficiente de correlación matemática (un valor cercano a `0.98 - 1.0` indica un contador global incremental predecible).
   * **`TIPO DE GENERACIÓN`**: Determina la implementación de la pila TCP/IP (`GLOBAL_INCREMENTAL`, `RANDOMIZED` o `PER_HOST_HASH`).
   * **`HUELLA CORRELACIÓN`**: Firma del sistema operativo inferida (ej: `LINUX_KERNEL_6.X_INCREMENTAL`).

   #### 🅲️ Tarjeta 3: Túneles VPN & Centrales FTTH (BAP/RIMA)
   * **`TÚNEL VPN / PROXY`**: Detecta si el tráfico está siendo encapsulado por WireGuard, OpenVPN, IPsec o proxies HTTP.
   * **`NIVEL DE CERTEZA`**: Porcentaje de confianza del análisis forense (ej: `92%`).
   * **`RETARDO L4 TCP / L7 APP`**: Compara la latencia a nivel de capa de transporte (SYN/ACK) frente a la capa de aplicación (HTTP GET).
   * **`DELTA LATENCIA (DISCREPANCIA)`**: Una discrepancia elevada revela la presencia de servidores intermediarios o túneles de cifrado intermedios.

---

## 8. Resolución de Problemas Frecuentes (Troubleshooting)

### P1: El mapa aparece en gris o no descarga las calles
* **Causa:** No hay conexión a Internet activa o el servidor de teselas de OpenStreetMap tardó en responder.
* **Solución:** Verifica que el móvil tenga datos móviles o Wi-Fi habilitados. Una vez que se descargan las teselas por primera vez, la app las guarda en caché local (`osmdroid/tiles`) para uso sin conexión.

### P2: El escáner Wi-Fi no muestra redes o solo muestra la red conectada
* **Causa:** En versiones recientes de Android (Android 10 a 15), el sistema operativo activa por defecto la limitación de escaneo Wi-Fi (*Wi-Fi Scan Throttling*).
* **Solución:** 
  1. Ve a los **Ajustes de Android** > **Opciones para desarrolladores**.
  2. Busca la opción **"Limitación de búsqueda de redes Wi-Fi"** (*Wi-Fi scan throttling*) y desactívala.
  3. Asegúrate de haber concedido el permiso de **Ubicación Precisa**.

### P3: El servicio de Wardriving se detiene al apagar la pantalla
* **Causa:** El optimizador de batería de fabricantes como Xiaomi (MIUI), Samsung (OneUI) o Huawei cierra servicios en segundo plano agresivamente.
* **Solución:**
  1. Ve a **Ajustes de la aplicación Rastreador**.
  2. En **Batería / Ahorro de energía**, selecciona **"Sin restricciones"** (*Unrestricted*).

### P4: Al pulsar "RASTREAR" da error "Fallo de comunicación con el motor Go"
* **Causa:** Se introdujo una dirección IP privada (ej: `192.168.1.1` o `127.0.0.1`) que no tiene ruta en las tablas BGP globales de Internet.
* **Solución:** Utiliza IPs públicas enrutables (ej: `2.139.25.3`, `8.8.8.8`) o pasa la dirección MAC de un router para triangulación L2.

---

## 9. Especificaciones de Seguridad y Privacidad

* **Cero Recolección Externa:** Toda la telemetría recolectada por los sensores y las consultas se procesa en el propio dispositivo local.
* **Almacenamiento Local Cifrado:** Las capturas y logs de wardriving se almacenan en la base de datos interna SQLite de la aplicación, sin compartir datos con servidores de terceros no autorizados.
* **Uso Ético:** Esta herramienta está destinada a auditorías autorizadas, análisis de infraestructura propia e investigación técnica de seguridad defensiva.

---

*Manual oficial mantenido para Rastreador Mobile en la rama `android`.*
