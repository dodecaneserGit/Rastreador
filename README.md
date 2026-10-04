# 🎯 Rastreador

**Motor de Geolocalización Activa por Multilateración y Desanonimización de Capas de Red (Go)**

`Rastreador` es una herramienta multiplataforma (macOS, Linux, Windows) diseñada para determinar la ubicación física de una dirección IP mediante el modelo matemático de restricciones **Constraint-Based Geolocation (CBG)**, física de propagación de fibra óptica ($c_{fibra} \approx 200.000\text{ km/s}$), reconocimiento BGP/ASN y análisis diferencial de retardo entre la capa de transporte (L4) y la capa de aplicación (L7) para desenmascarar nodos ocultos tras VPNs o Proxies.

---

## ⚡ Características Principales

1. **Reconocimiento BGP y Clasificación de Red**:
   - Consultas DNS directas a Team Cymru para resolución instantánea de ASN, ISP y Organización BGP sin depender de APIs lentas o con rate-limit.
   - Detección de datacenters, clouds públicas (AWS, GCP, Azure, Hetzner, DigitalOcean, OVH) y proveedores comerciales de VPN.
   - Identificación de códigos IATA de metro/aeropuertos en registros PTR inversos.

2. **Multilateración Activa CBG (Constraint-Based Geolocation)**:
   - Sondas de retardo (RTT) ultra-rápidas mediante TCP SYN y ACK en puertos estándar.
   - Cálculo del radio máximo físico según la velocidad de la luz en vidrio ($d_{max} = \frac{RTT \times 200}{2}$).
   - Solver geométrico WGS84 (Haversine) para estimar las coordenadas del host y su radio de confianza.

3. **Análisis de Túneles y Diferencial de Latencia (L4 vs L7)**:
   - Medición del $\Delta RTT = RTT_{L7} - RTT_{L4}$ para acotar la distancia física real de clientes tras nodos de salida VPN.
   - Detección de sobrecarga de encapsulación (MTU/MSS).

4. **Visualización y Reportes**:
   - Generación automática de mapas interactivos en HTML basados en **Leaflet.js** con modo oscuro, circunferencias de restricción y tarjetas HUD.
   - Exportación de reportes detallados en formato JSON estructurado.

---

## 🚀 Compilación e Instalación

Requiere **Go 1.22+**:

```bash
git clone https://github.com/dodecaneser/rastreador.git
cd Rastreador
go build -o bin/rastreador ./cmd/rastreador
```

---

## 📖 Uso y Ejemplos

### 1. Escaneo y Geolocalización Básica
```bash
./bin/rastreador -ip 1.1.1.1 -map map.html -out report.json
```

### 2. Escaneo con Análisis de Diferencial de Túnel L4/L7
```bash
./bin/rastreador -ip 198.51.100.4 -l7-url "https://target-domain.com/ping" -map vpn_map.html
```

### 3. Salida en formato JSON crudo para automatizaciones
```bash
./bin/rastreador -ip 8.8.8.8 -json
```

---

## 🛠️ Opciones del CLI

| Flag | Tipo | Descripción |
| :--- | :--- | :--- |
| `-ip` | `string` | Dirección IP objetivo a analizar y geolocalizar (requerido). |
| `-ports` | `string` | Lista de puertos TCP para sondas RTT (defecto: `80,443,22,53,8080`). |
| `-samples` | `int` | Número de paquetes de prueba por sonda para filtrar ruido (defecto: `5`). |
| `-l7-url` | `string` | URL HTTP/S en el objetivo para medir la diferencia de latencia L4 vs L7. |
| `-map` | `string` | Ruta de salida para el mapa HTML interactivo (defecto: `map_result.html`). |
| `-out` | `string` | Ruta de guardado para el informe estructurado en JSON. |
| `-json` | `bool` | Imprime únicamente JSON por stdout. |
