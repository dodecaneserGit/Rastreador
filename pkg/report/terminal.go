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

	fmt.Println(strings.Repeat("═", 72) + "\n")
}
