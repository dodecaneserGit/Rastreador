package report

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dodecaneser/rastreador/pkg/multilat"
	"github.com/dodecaneser/rastreador/pkg/recon"
	"github.com/dodecaneser/rastreador/pkg/tunnel"
)

// FullScanReport aggregates all reconnaissance, multilateration and tunnel inspection findings
type FullScanReport struct {
	TargetIP        string                         `json:"target_ip"`
	Timestamp       string                         `json:"timestamp"`
	ReconInfo       *recon.IPInfo                  `json:"recon_info"`
	Multilateration *multilat.MultilaterationResult `json:"multilateration,omitempty"`
	TunnelAnalysis  *tunnel.TunnelAnalysis         `json:"tunnel_analysis,omitempty"`
}

// GenerateHTMLMap creates an interactive Leaflet.js visualization of landmarks, circles, and estimated point
// Uses 100% free, public OpenStreetMap tiles that DO NOT require any API key.
func GenerateHTMLMap(report *FullScanReport, outputPath string) error {
	if report.Multilateration == nil {
		return fmt.Errorf("no multilateration data available to generate map")
	}

	est := report.Multilateration.EstimatedPoint
	landmarksJSON, _ := json.Marshal(report.Multilateration.UsedLandmarks)
	reconJSON, _ := json.Marshal(report.ReconInfo)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Rastreador - Geolocalización de IP [%s]</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <style>
        body { margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; }
        #map { height: 100vh; width: 100%%; }
        .dashboard-card {
            position: absolute; top: 20px; left: 20px; z-index: 1000;
            background: rgba(15, 23, 42, 0.94); backdrop-filter: blur(12px);
            padding: 20px; border-radius: 12px; border: 1px solid #334155;
            max-width: 380px; box-shadow: 0 12px 30px rgba(0,0,0,0.6);
        }
        .title { font-size: 1.25rem; font-weight: 700; color: #38bdf8; margin-bottom: 8px; display: flex; align-items: center; gap: 8px; }
        .badge { display: inline-block; padding: 4px 10px; border-radius: 6px; font-size: 0.75rem; font-weight: 700; text-transform: uppercase; margin-bottom: 14px; letter-spacing: 0.5px; }
        .badge-vpn { background: #dc2626; color: #fff; }
        .badge-cloud { background: #ea580c; color: #fff; }
        .badge-direct { background: #16a34a; color: #fff; }
        .stat-row { display: flex; justify-content: space-between; margin-bottom: 7px; font-size: 0.85rem; border-bottom: 1px solid #1e293b; padding-bottom: 5px; }
        .stat-label { color: #94a3b8; }
        .stat-val { font-weight: 600; color: #f8fafc; text-align: right; }
        .gmaps-btn {
            display: block; width: 100%%; box-sizing: border-box; margin-top: 14px; padding: 8px 12px;
            background: #0284c7; color: #fff; text-align: center; text-decoration: none;
            border-radius: 6px; font-size: 0.85rem; font-weight: 600; transition: background 0.2s;
        }
        .gmaps-btn:hover { background: #0369a1; }
    </style>
</head>
<body>
    <div class="dashboard-card">
        <div class="title">🎯 Rastreador IP Matrix</div>
        <div class="badge %s">%s</div>
        <div class="stat-row"><span class="stat-label">IP Objetivo:</span><span class="stat-val">%s</span></div>
        <div class="stat-row"><span class="stat-label">ASN / Organización:</span><span class="stat-val">AS%d (%s)</span></div>
        <div class="stat-row"><span class="stat-label">País / ISP:</span><span class="stat-val">%s / %s</span></div>
        <div class="stat-row"><span class="stat-label">Coord. Estimadas:</span><span class="stat-val">%.4f, %.4f</span></div>
        <div class="stat-row"><span class="stat-label">Radio Confianza:</span><span class="stat-val">±%.1f km</span></div>
        <div class="stat-row"><span class="stat-label">Sondas Activas:</span><span class="stat-val">%d Nodos</span></div>
        <a class="gmaps-btn" href="https://www.google.com/maps?q=%.4f,%.4f" target="_blank">Abrir en Google Maps ↗</a>
    </div>
    <div id="map"></div>

    <script>
        const estLat = %f;
        const estLon = %f;
        const landmarks = %s;
        const reconData = %s;

        // Initialize Map centered on estimated coordinates
        const map = L.map('map').setView([estLat, estLon], 5);

        // 100%% Free OpenStreetMap Tile Layer without API key requirements
        L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            maxZoom: 19
        }).addTo(map);

        // Target Estimated Marker
        const targetIcon = L.divIcon({
            className: 'target-marker',
            html: '<div style="background-color:#ef4444;width:20px;height:20px;border-radius:50%%;border:3px solid #fff;box-shadow:0 0 15px #ef4444;"></div>',
            iconSize: [26, 26],
            iconAnchor: [13, 13]
        });

        const targetMarker = L.marker([estLat, estLon], {icon: targetIcon}).addTo(map)
            .bindPopup("<b>🎯 Posición Física Estimada</b><br>Lat: " + estLat.toFixed(4) + "<br>Lon: " + estLon.toFixed(4) + "<br>Confianza: ±" + %f.toFixed(1) + " km")
            .openPopup();

        // Target Confidence Circle
        L.circle([estLat, estLon], {
            color: '#ef4444',
            fillColor: '#ef4444',
            fillOpacity: 0.18,
            weight: 2,
            radius: %f * 1000
        }).addTo(map);

        // Landmark Probes and Constraint Circles
        landmarks.forEach(lm => {
            if (lm.location && lm.location.lat) {
                // Landmark Marker
                L.circleMarker([lm.location.lat, lm.location.lon], {
                    radius: 7,
                    color: '#0284c7',
                    fillColor: '#38bdf8',
                    fillOpacity: 0.9,
                    weight: 2
                }).addTo(map).bindPopup("<b>📍 " + lm.name + " (" + lm.city + ")</b><br>RTT Mínimo: " + lm.min_rtt_ms.toFixed(2) + " ms<br>Radio Máximo CBG: " + lm.max_radius_km.toFixed(0) + " km");

                // Constraint Circle
                L.circle([lm.location.lat, lm.location.lon], {
                    color: '#0284c7',
                    weight: 1,
                    dashArray: '5, 8',
                    fillColor: '#38bdf8',
                    fillOpacity: 0.05,
                    radius: lm.max_radius_km * 1000
                }).addTo(map);
            }
        });
    </script>
</body>
</html>`,
		report.TargetIP,
		getBadgeClass(report.ReconInfo),
		report.ReconInfo.Confidence,
		report.TargetIP,
		report.ReconInfo.ASN,
		report.ReconInfo.ASOrg,
		report.ReconInfo.Country,
		report.ReconInfo.ISP,
		est.Lat,
		est.Lon,
		report.Multilateration.ConfidenceKm,
		len(report.Multilateration.UsedLandmarks),
		est.Lat,
		est.Lon,
		est.Lat,
		est.Lon,
		string(landmarksJSON),
		string(reconJSON),
		report.Multilateration.ConfidenceKm,
		report.Multilateration.ConfidenceKm,
	)

	return os.WriteFile(outputPath, []byte(html), 0644)
}

// OpenInBrowser opens the specified file path or URL in the system default web browser
func OpenInBrowser(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err == nil {
		filePath = absPath
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filePath)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", filePath)
	}

	return cmd.Start()
}

func getBadgeClass(info *recon.IPInfo) string {
	if info.IsVPN || info.IsProxy {
		return "badge-vpn"
	}
	if info.IsCloud {
		return "badge-cloud"
	}
	return "badge-direct"
}
