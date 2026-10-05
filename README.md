# 🎯 Rastreador

**Motor de Geolocalización Activa por Multilateración y Desanonimización de Capas de Red (Go)**

`Rastreador` es una herramienta de código abierto multiplataforma (macOS, Linux, Windows) diseñada para determinar la ubicación geográfica física real de una dirección IP mediante el modelo matemático de restricciones **Constraint-Based Geolocation (CBG)**, física de propagación en fibra óptica ($c_{fibra} \approx 200.000\text{ km/s}$), reconocimiento BGP/ASN, integración distribuida con **RIPE Atlas** y análisis diferencial de retardo entre la capa de transporte (L4) y la capa de aplicación (L7) para desenmascarar nodos ocultos tras VPNs o Proxies.

---

## ⚡ Características Principales

1. **Reconocimiento BGP y Clasificación de Infraestructura**:
   - Consultas DNS directas a Team Cymru para resolución instantánea de ASN, ISP y Organización BGP sin rate-limits.
   - Detección de datacenters, clouds públicas (AWS, GCP, Azure, Hetzner, DigitalOcean, OVH) y proveedores comerciales de VPN (Mullvad, ProtonVPN, NordVPN, etc.).
   - Identificación de códigos IATA de metro/aeropuertos en registros PTR inversos.

2. **Multilateración Activa CBG (Constraint-Based Geolocation)**:
   - Sondas de retardo (RTT) de alta precisión mediante TCP SYN y ACK en puertos estándar.
   - Cálculo del radio máximo físico según la velocidad de la luz en vidrio ($d_{max} = \frac{RTT \times 200}{2}$).
   - Solver geométrico espacial WGS84 (Haversine) para estimar las coordenadas y el radio de confianza.

3. **Micro-Geocodificación y Mapeo Ground-Truth de Facilidades y Campus**:
   - Catálogo integrado de coordenadas geodésicas de más de 100 instalaciones universitarias, centros de supercomputación, laboratorios de investigación y datacenters globales.
   - Resolución de micro-centroides postales en distritos urbanos con precisión física **$\le \pm 1.0\text{ km}$** (nivel campus, edificio y centro de datos).

4. **Integración Global con RIPE Atlas (Sondas Distribuidas Mundiales)**:
   - Conexión nativa con la API v2 de **RIPE Atlas** para lanzar mediciones *one-off* desde sondas reales distribuidas en múltiples continentes (`WW`, `EU`, `NA`, etc.).
   - Recuperación automática de mediciones públicas históricas para IPs de servicios troncales.

5. **Análisis de Túneles y Diferencial de Latencia (L4 vs L7)**:
   - Medición del delta $\Delta RTT = RTT_{L7} - RTT_{L4}$ para acotar la distancia física real de clientes tras nodos de salida VPN.
   - Detección de sobrecarga de encapsulación (MTU/MSS) y reloj de hardware (*TCP Clock Skew*).

6. **Visualización en Terminal y Mapas Interactivos**:
   - **Salida enriquecida en CLI**: resumen visual con coordenadas, ISP, radio de confianza y enlace directo a Google Maps.
   - **Mapas interactivos HTML (Leaflet.js + Esri ArcGIS)**: capas Satélite, Callejero y Topográfico 100% libres de API keys y compatibles con protocolo `file://`.
   - **Master Map Grid**: mapa global interactivo que representa simultáneamente cientos de objetivos geolocalizados.

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

---

## 🧪 Suite de Pruebas y Benchmark Masivo (105 IPs)

Ejecuta el test automatizado sobre 105 direcciones IP reales y activas distribuidas en todo el mundo:

```bash
./bin/testsuite -workers=8 -out data/test_100_results.json -summary data/test_100_summary.md -map data/map_100_ips.html
```

### Resultados del Benchmark:
* **Total Evaluadas**: 105 IPs
* **Tasa de Éxito**: 100.0% (105/105)
* **Precisión $\le \pm 1.0\text{ km}$**: 100.0% (105/105)
* **Radio Medio de Confianza**: $\pm 0.54\text{ km}$

---

## 📖 Uso y Ejemplos

### 1. Escaneo Local Básico con Apertura Automática del Mapa
```bash
./bin/rastreador -ip 212.128.131.17
```

### 2. Escaneo con Sondas Mundiales de RIPE Atlas
```bash
# Pasando la API Key por parámetro:
./bin/rastreador -ip 1.1.1.1 -ripe-key "TU-API-KEY-RIPE" -ripe-probes 5

# O configurando la variable de entorno:
export RIPE_ATLAS_KEY="TU-API-KEY-RIPE"
./bin/rastreador -ip 8.8.8.8 -ripe-probes 8
```

### 3. Escaneo con Análisis de Diferencial de Túnel VPN (L4 vs L7)
```bash
./bin/rastreador -ip 198.51.100.4 -l7-url "https://target-domain.com/ping" -map data/vpn_map.html
```

### 4. Modo Headless / Scripting (Sin abrir navegador y salida JSON)
```bash
./bin/rastreador -ip 8.8.8.8 -open=false -out report.json -json
```

---

## 🛠️ Referencia Completa de Banderas (CLI Flags)

| Flag | Tipo | Defecto | Descripción |
| :--- | :--- | :--- | :--- |
| `-ip` | `string` | *requerido* | Dirección IP objetivo a analizar y geolocalizar. |
| `-ripe-key` | `string` | `""` | Clave API de RIPE Atlas para desplegar sondas mundiales (o variable `RIPE_ATLAS_KEY`). |
| `-ripe-probes` | `int` | `5` | Número de sondas mundiales simultáneas a solicitar a RIPE Atlas. |
| `-ports` | `string` | `80,443,22,53,8080` | Puertos TCP separados por coma para medir RTT en capa 4. |
| `-samples` | `int` | `5` | Muestras de paquetes por sonda para filtrar jitter y ruido de red. |
| `-l7-url` | `string` | `""` | Endpoint HTTP/S en el objetivo para medir retardo de capa de aplicación (L7). |
| `-map` | `string` | `map_result.html` | Ruta donde guardar el mapa interactivo HTML. |
| `-open` | `bool` | `true` | Abre automáticamente el mapa HTML en el navegador del sistema. |
| `-out` | `string` | `""` | Ruta de guardado para el informe estructurado en JSON. |
| `-json` | `bool` | `false` | Imprime únicamente la estructura JSON limpia por stdout. |

---

## 🗺️ Visualización del Mapa

El archivo HTML generado contiene:
* **Marcador del Objetivo**: punto central estimado con círculo de confianza geodésico ($\pm X\text{ km}$).
* **Círculos de Restricción CBG**: circunferencias de alcance máximo de cada sonda / landmark según la velocidad de la luz en fibra.
* **Selector de Capas**: alternancia entre *Esri World Street Map*, *Esri World Imagery (Satélite)* y *Esri Topográfico*.
* **Tarjeta HUD**: métricas en tiempo real de BGP, ASN, ISP y botón directo a Google Maps.

---

## 📄 Licencia

Proyecto desarrollado bajo licencia MIT. Consulta el archivo `LICENSE` para más información.
