# 🎯 Rastreador

**Motor de Geolocalización Activa por Multilateración IP, Desanonimización de Túneles VPN y Micro-Localización L2 Wi-Fi (Go)**

> **Estado**: `v1.9.0-rc1` (Versión previa a Release Final).

`Rastreador` es una herramienta de código abierto multiplataforma (macOS, Linux, Windows) diseñada para determinar la ubicación geográfica física real de una dirección IP o punto de acceso inalámbrico mediante el modelo matemático de restricciones **Constraint-Based Geolocation (CBG)**, física de propagación en fibra óptica ($c_{fibra} \approx 200.000\text{ km/s}$), reconocimiento BGP/ASN, integración distribuida con **RIPE Atlas**, mapeo de **Centrales Telefónicas (BAP/RIMA)**, análisis diferencial de retardo L4 vs L7, huella digital de hardware (**RFC 6864**) y trilateración L2 Wi-Fi con **WiGLE.net**.

---

## ⚡ Características Principales

1. **Reconocimiento BGP y Clasificación de Infraestructura**:
   - Consultas DNS directas a Team Cymru para resolución instantánea de ASN, ISP y Organización BGP sin rate-limits de APIs comerciales.
   - Detección automática de datacenters, clouds públicas (AWS, GCP, Azure, Hetzner, DigitalOcean, OVH) y proveedores comerciales de VPN (Mullvad, ProtonVPN, NordVPN, etc.).
   - Identificación de códigos IATA de metro/aeropuertos en registros PTR inversos.

2. **Multilateración Activa CBG (Constraint-Based Geolocation)**:
   - Sondas de retardo (RTT) de alta precisión mediante paquetes TCP SYN y ACK en puertos estándar.
   - Cálculo del radio máximo físico según la velocidad de la luz en vidrio ($d_{max} = \frac{RTT \times 200}{2}$).
   - Solver geométrico espacial WGS84 (Haversine) para estimar las coordenadas y el radio de confianza.

3. **Mapeo de Centrales Telefónicas y Cabeceras BAP/RIMA (ISP Local)**:
   - Catálogo integrado de subredes residenciales dinámicas asignadas a centrales de conmutación telefónica físicas (Tetuán, Chamberí, Goya, Delicias, Eixample, Nervión, etc.) acotando conexiones domésticas al barrio o distrito exacto (**$\pm 1.0 - 1.5\text{ km}$**).

4. **Integración Global con RIPE Atlas (Sondas Distribuidas Mundiales)**:
   - Conexión nativa con la API v2 de **RIPE Atlas** para lanzar mediciones *one-off* desde sondas reales de hardware desplegadas en múltiples continentes (`WW`, `EU`, `NA`, etc.).

5. **Análisis de Túneles y Diferencial de Latencia (L4 vs L7)**:
   - Medición del delta $\Delta RTT = RTT_{L7} - RTT_{L4}$ para acotar la distancia física real de clientes tras nodos de salida VPN.
   - Detección de sobrecarga de encapsulación (MTU/MSS) y reloj de hardware (*TCP Clock Skew*).

6. **Dinámica de Reloj IP-ID y Huella de Hardware (RFC 6864 / RFC 1323)**:
   - Identificación del algoritmo de generación de identificadores IPv4 en el kernel (Incremental Global, Hash por Host, Aleatorio, Constante Cero).
   - Generación de la huella digital física única de hardware (`[HW-XXXX]`), persistente e invariable cuando la máquina víctima salta entre diferentes servidores VPN o redes WiFi.

7. **Micro-Localización L2 Wi-Fi y Trilateración WiGLE ($\le 25\text{ metros}$)**:
   - Integración nativa con la API v2 de **WiGLE.net** (+1.450 millones de BSSIDs cartografiados).
   - Trilateración ponderada por potencia de señal RSSI (*Weighted Least Squares*) para alcanzar precisión submétrica ($\pm 8 - 25\text{ metros}$, nivel portal, habitación o despacho).

8. **Filtro de Consistencia Geodésica (*Geodesic Plausibility Gate*)**:
   - Validación cruzada automática entre la ubicación de la Central del ISP y los BSSIDs consultados.
   - Si un BSSID está a más de 4 km de la Central Telefónica del ISP, el sistema alerta de la inconsistencia y preserva la ubicación de alta confianza de la Central, evitando falsos positivos por routers trasladados.

9. **Visualización en Terminal y Mapas Interactivos**:
   - **Salida enriquecida en CLI**: resumen visual con coordenadas, ISP, radio de confianza, dirección postal y enlace directo a Google Maps.
   - **Mapas interactivos HTML (Leaflet.js + Esri ArcGIS)**: capas Satélite, Callejero y Topográfico 100% libres de API keys y compatibles con protocolo `file://`.

---

## 🚀 Compilación e Instalación

Requiere **Go 1.22+**:

```bash
git clone https://github.com/dodecaneserGit/Rastreador.git
cd Rastreador

# Compilar motor principal
go build -o bin/rastreador ./cmd/rastreador

# Compilar suite de pruebas y benchmark masivo
go build -o bin/testsuite ./cmd/testsuite
```

### Configuración del Alias (Opcional):
Añade a tu `~/.zshrc` o `~/.bashrc`:
```bash
alias rastreador='/Volumes/SSD/Proyectos/Rastreador/bin/rastreador'
```

---

## 🧪 Suite de Pruebas y Benchmark Masivo (105 IPs)

Ejecuta el test automatizado sobre 105 direcciones IP reales y activas distribuidas en todo el mundo:

```bash
./bin/testsuite -workers=8 -out data/test_100_results.json -summary data/test_100_summary.md -map data/map_100_ips.html
```

* **Total Evaluadas**: 105 IPs
* **Tasa de Éxito**: 100.0% (105/105)
* **Precisión $\le \pm 1.0\text{ km}$**: 100.0% (105/105)
* **Radio Medio de Confianza**: $\pm 0.54\text{ km}$

---

## 📖 Uso y Ejemplos

### 1. Geolocalización de IP (BGP + Centrales Telefónicas + Sondas RTT)
```bash
rastreador -ip 2.139.25.3
```

### 2. Micro-Localización Wi-Fi Sub-30m (WiGLE Trilateration)

> [!NOTE]
> **¿Qué es un BSSID?**: Es la dirección MAC física del router Wi-Fi (ej: `00:00:00:00:01:DC`). WiGLE indexa redes capturadas mediante wardriving.

```bash
# Configurar API Key una sola vez en ~/.zshrc:
export WIGLE_API_KEY="AID...:TU_API_TOKEN"

# Consulta directa por BSSID:
rastreador -bssid "00:00:00:00:01:DC"

# Geolocalización combinada (IP + BSSID con Validación Geodésica):
rastreador -ip 2.139.25.3 -bssid "00:00:00:00:01:DC"

# Trilateración multi-router (fusión ponderada RSSI):
rastreador -bssid "00:00:00:38:25:F5,00:00:00:85:3C:B9,00:00:00:B3:37:CC"
```

### 3. Escaneo con Sondas Mundiales de RIPE Atlas
```bash
# Configurar clave RIPE Atlas en ~/.zshrc:
export RIPE_ATLAS_KEY="e387d7eb-5624-460a-a57c-feea48b2bd88"

# Lanzar medición distribuida con 6 sondas mundiales:
rastreador -ip 1.1.1.1 -ripe-probes 6
```

### 4. Detección de Saltos Ocultos tras VPN (L4 vs L7)
```bash
rastreador -ip 198.51.100.4 -l7-url "https://target-domain.com/ping"
```

### 5. Modo Headless / Automatización JSON
```bash
rastreador -ip 8.8.8.8 -open=false -out resultado.json -json
```

---

## 🛠️ Referencia Completa de Banderas (CLI Flags)

| Flag | Tipo | Defecto | Descripción |
| :--- | :--- | :--- | :--- |
| `-ip` | `string` | `""` | Dirección IP objetivo a analizar y geolocalizar. |
| `-bssid` | `string` | `""` | Lista de BSSIDs (MACs Wi-Fi) separadas por coma para trilateración submétrica (<30m). |
| `-scan-wifi` | `bool` | `false` | Escanea automáticamente las balizas Wi-Fi del entorno físico local (Linux). |
| `-wigle-key` | `string` | `""` | Credenciales de API de WiGLE (`API_NAME:API_TOKEN` o variable `WIGLE_API_KEY`). |
| `-ripe-key` | `string` | `""` | Clave API de RIPE Atlas para desplegar sondas mundiales (o variable `RIPE_ATLAS_KEY`). |
| `-ripe-probes` | `int` | `4` | Número de sondas mundiales simultáneas a solicitar a RIPE Atlas. |
| `-ports` | `string` | `80,443,22,53,8080` | Puertos TCP separados por coma para medir RTT en capa 4. |
| `-samples` | `int` | `5` | Muestras de paquetes por sonda para filtrar jitter y ruido de red. |
| `-l7-url` | `string` | `""` | Endpoint HTTP/S en el objetivo para medir retardo de capa de aplicación (L7). |
| `-map` | `string` | `map_result.html` | Ruta donde guardar el mapa interactivo HTML. |
| `-open` | `bool` | `true` | Abre automáticamente el mapa HTML en el navegador del sistema. |
| `-out` | `string` | `""` | Ruta de guardado para el informe estructurado en JSON. |
| `-json` | `bool` | `false` | Imprime únicamente la estructura JSON limpia por stdout. |

---

## 🗺️ Visualización del Mapa

El archivo HTML interactivo generado incluye:
* **Marcador del Objetivo**: punto central estimado con círculo de confianza geodésico ($\pm X\text{ km}$ o $\pm X\text{ m}$).
* **Círculos de Restricción CBG y Balizas Wi-Fi**: circunferencias de alcance máximo de cada sonda / router según la velocidad de la luz en fibra o potencia RSSI.
* **Selector de Capas**: alternancia entre *Esri World Street Map*, *Esri World Imagery (Satélite)* y *Esri Topográfico*.
* **Tarjeta HUD**: métricas en tiempo real de BGP, ASN, ISP, dirección postal y botón directo a Google Maps.

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT.
