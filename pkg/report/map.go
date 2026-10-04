package report

import (
	"encoding/json"
	"fmt"
	"os"

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
func GenerateHTMLMap(report *FullScanReport, outputPath string) error {
	if report.Multilateration == nil {
		return fmt.Errorf("no multilateration data available to generate map")
	}

	est := report.Multilateration.EstimatedPoint
	landmarksJSON, _ := json.Marshal(report.Multilateration.UsedLandmarks)
	reconJSON, _ := json.Marshal(report.ReconInfo)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
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
            background: rgba(15, 23, 42, 0.92); backdrop-filter: blur(10px);
            padding: 20px; border-radius: 12px; border: 1px solid #334155;
            max-width: 360px; box-shadow: 0 10px 25px rgba(0,0,0,0.5);
        }
        .title { font-size: 1.25rem; font-weight: 700; color: #38bdf8; margin-bottom: 8px; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 6px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; margin-bottom: 12px; }
        .badge-vpn { background: #dc2626; color: #fff; }
        .badge-cloud { background: #ea580c; color: #fff; }
        .badge-direct { background: #16a34a; color: #fff; }
        .stat-row { display: flex; justify-content: space-between; margin-bottom: 6px; font-size: 0.85rem; border-bottom: 1px solid #1e293b; padding-bottom: 4px; }
        .stat-label { color: #94a3b8; }
        .stat-val { font-weight: 600; color: #f8fafc; }
    </style>
</head>
<body>
    <div class="dashboard-card">
        <div class="title">🎯 Rastreador IP Matrix</div>
        <div class="badge %s">%s</div>
        <div class="stat-row"><span class="stat-label">IP Objetivo:</span><span class="stat-val">%s</span></div>
        <div class="stat-row"><span class="stat-label">ASN / Org:</span><span class="stat-val">AS%d (%s)</span></div>
        <div class="stat-row"><span class="stat-label">Coord. Estimadas:</span><span class="stat-val">%.4f, %.4f</span></div>
        <div class="stat-row"><span class="stat-label">Radio Confianza:</span><span class="stat-val">±%.1f km</span></div>
        <div class="stat-row"><span class="stat-label">Nodos Probing:</span><span class="stat-val">%d Nodos</span></div>
    </div>
    <div id="map"></div>

    <script>
        const estLat = %f;
        const estLon = %f;
        const landmarks = %s;
        const reconData = %s;

        const map = L.map('map').setView([estLat, estLon], 5);

        L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
            attribution: '&copy; OpenStreetMap contributors &copy; CARTO',
            subdomains: 'abcd',
            maxZoom: 19
        }).addTo(map);

        // Add Target Estimated Marker
        const targetIcon = L.divIcon({
            className: 'target-marker',
            html: '<div style="background-color:#ef4444;width:18px;height:18px;border-radius:50%%;border:3px solid #fff;box-shadow:0 0 15px #ef4444;"></div>',
            iconSize: [24, 24],
            iconAnchor: [12, 12]
        });

        const targetMarker = L.marker([estLat, estLon], {icon: targetIcon}).addTo(map)
            .bindPopup("<b>🎯 Posición Física Estimada</b><br>Lat: " + estLat.toFixed(4) + "<br>Lon: " + estLon.toFixed(4) + "<br>Confianza: ±" + %f.toFixed(1) + " km")
            .openPopup();

        // Target Confidence Circle
        L.circle([estLat, estLon], {
            color: '#ef4444',
            fillColor: '#ef4444',
            fillOpacity: 0.15,
            radius: %f * 1000
        }).addTo(map);

        // Add Landmark Probes and Radii
        landmarks.forEach(lm => {
            if (lm.location && lm.location.lat) {
                // Landmark Marker
                L.circleMarker([lm.location.lat, lm.location.lon], {
                    radius: 6,
                    color: '#38bdf8',
                    fillColor: '#0284c7',
                    fillOpacity: 0.9,
                    weight: 2
                }).addTo(map).bindPopup("<b>📍 " + lm.name + " (" + lm.city + ")</b><br>RTT Mínimo: " + lm.min_rtt_ms.toFixed(2) + " ms<br>Radio Máximo CBG: " + lm.max_radius_km.toFixed(0) + " km");

                // Constraint Circle
                L.circle([lm.location.lat, lm.location.lon], {
                    color: '#0284c7',
                    weight: 1,
                    dashArray: '4, 8',
                    fillColor: '#38bdf8',
                    fillOpacity: 0.04,
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
		est.Lat,
		est.Lon,
		report.Multilateration.ConfidenceKm,
		len(report.Multilateration.UsedLandmarks),
		est.Lat,
		est.Lon,
		string(landmarksJSON),
		string(reconJSON),
		report.Multilateration.ConfidenceKm,
		report.Multilateration.ConfidenceKm,
	)

	return os.WriteFile(outputPath, []byte(html), 0644)
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
