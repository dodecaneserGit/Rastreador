package report

import (
	"fmt"
	"strings"
)

// PrintTerminalSummary prints a comprehensive, visual report directly to stdout
func PrintTerminalSummary(rep *FullScanReport) {
	fmt.Println("\n" + strings.Repeat("═", 72))
	fmt.Println("                  📊 RESULTADOS DEL RASTREO Y GEOLOCALIZACIÓN")
	fmt.Println(strings.Repeat("═", 72))

	recon := rep.ReconInfo
	fmt.Println("  🌐 INFORMACIÓN DE RED E INFRAESTRUCTURA:")
	fmt.Printf("    • IP Objetivo         : \033[1;36m%s\033[0m\n", rep.TargetIP)
	if recon.Hostname != "" {
		fmt.Printf("    • Hostname (PTR)      : %s\n", recon.Hostname)
	}
	fmt.Printf("    • Sistema Autónomo    : AS%d (%s)\n", recon.ASN, recon.ASOrg)
	fmt.Printf("    • Proveedor / ISP     : %s\n", recon.ISP)
	if recon.Country != "" || recon.City != "" {
		fmt.Printf("    • País / Ciudad Reg.  : %s / %s\n", recon.Country, recon.City)
	}
	if recon.Facility != "" {
		fmt.Printf("    • Campus / Facilidad  : \033[1;32m%s\033[0m\n", recon.Facility)
	}
	if recon.AirportCode != "" {
		fmt.Printf("    • Metro Code (IATA)   : %s\n", recon.AirportCode)
	}
	fmt.Printf("    • Categoría de Red    : \033[1;33m[%s]\033[0m\n", recon.Confidence)

	if rep.Multilateration != nil && rep.Multilateration.EstimatedPoint.Lat != 0 {
		est := rep.Multilateration.EstimatedPoint
		fmt.Println("\n  📍 UBICACIÓN FÍSICA ESTIMADA (MULTILATERACIÓN ACTIVA):")
		fmt.Printf("    • Coordenadas (WGS84) : \033[1;32m%.4f, %.4f\033[0m\n", est.Lat, est.Lon)
		fmt.Printf("    • Radio de Confianza  : ±%.1f km\n", rep.Multilateration.ConfidenceKm)
		fmt.Printf("    • Enlace Google Maps  : \033[4;34mhttps://www.google.com/maps/search/?api=1&query=%.4f,%.4f\033[0m\n", est.Lat, est.Lon)
		fmt.Printf("    • Nodos / Hops        : %d landmarks activos\n", len(rep.Multilateration.UsedLandmarks))

		if len(rep.Multilateration.UsedLandmarks) > 0 {
			fmt.Println("    • Desglose por Sonda  :")
			for _, lm := range rep.Multilateration.UsedLandmarks {
				fmt.Printf("      - \033[1m%-24s\033[0m (%-10s): RTT mín = %6.2f ms | Radio CBG = %5.0f km\n",
					lm.Name, lm.City, lm.MinRTT, lm.MaxRadius)
			}
		}
	}

	if rep.TunnelAnalysis != nil && len(rep.TunnelAnalysis.Findings) > 0 {
		fmt.Println("\n  🛡️  ANÁLISIS DE TÚNELES Y ENCAPSULACIÓN (VPN/PROXY):")
		for _, f := range rep.TunnelAnalysis.Findings {
			fmt.Printf("    • %s\n", f)
		}
	}

	if rep.IPIDAnalysis != nil && len(rep.IPIDAnalysis.Findings) > 0 {
		ipidRes := rep.IPIDAnalysis
		fmt.Println("\n  ⏱️  DINÁMICA DE RELOJ IP-ID Y VELOCIDAD DE PAQUETES (RFC 6864):")
		fmt.Printf("    • Algoritmo de Kernel : \033[1;33m[%s]\033[0m\n", ipidRes.GenerationType)
		if ipidRes.VelocityPacketsPerSec > 0 {
			fmt.Printf("    • Velocidad de Emisión: \033[1;36m%.1f paquetes/segundo\033[0m (Linealidad R² = %.3f)\n", ipidRes.VelocityPacketsPerSec, ipidRes.LinearityScore)
		}
		if ipidRes.ClockFrequencyHz > 0 {
			fmt.Printf("    • Frecuencia TCP TS   : \033[1;32m%.0f Hz\033[0m\n", ipidRes.ClockFrequencyHz)
		}
		if ipidRes.CorrelationFingerprint != "" {
			fmt.Printf("    • Huella de Hardware  : \033[1;35m[HW-%s]\033[0m (Persistente ante salto de VPN)\n", ipidRes.CorrelationFingerprint)
		}
		for _, f := range ipidRes.Findings {
			fmt.Printf("    • %s\n", f)
		}
	}

	if rep.WiFiTriangulation != nil && rep.WiFiTriangulation.ResolvedCount > 0 {
		wifi := rep.WiFiTriangulation
		fmt.Println("\n  📶 MICRO-LOCALIZACIÓN L2 WI-FI (WIGLE TRILATERATION):")
		fmt.Printf("    • Coordenadas Baliza  : \033[1;32m%.6f, %.6f\033[0m\n", wifi.EstimatedPoint.Lat, wifi.EstimatedPoint.Lon)
		fmt.Printf("    • Precisión Sub-30m   : \033[1;32m±%.1f metros\033[0m (Nivel Habitación / Portal)\n", wifi.PrecisionM)
		if wifi.StreetAddress != "" {
			fmt.Printf("    • Dirección Postal    : \033[1;36m%s\033[0m\n", wifi.StreetAddress)
		}
		fmt.Printf("    • Balizas Resueltas   : %d / %d BSSIDs procesados\n", wifi.ResolvedCount, wifi.TotalBeacons)
		for _, n := range wifi.Networks {
			if n.Resolved {
				fmt.Printf("      - \033[1m%s\033[0m (%-16s): Lat/Lon: %.6f, %.6f | RSSI: %d dBm | %s\n",
					n.BSSID, n.SSID, n.Lat, n.Lon, n.RSSI, n.Road)
			}
		}
	}

	fmt.Println(strings.Repeat("═", 72) + "\n")
}
