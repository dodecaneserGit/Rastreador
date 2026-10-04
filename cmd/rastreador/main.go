package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dodecaneser/rastreador/pkg/multilat"
	"github.com/dodecaneser/rastreador/pkg/recon"
	"github.com/dodecaneser/rastreador/pkg/report"
	"github.com/dodecaneser/rastreador/pkg/tunnel"
)

const banner = `
  ██████╗  █████╗ ███████╗████████╗██████╗ ███████╗ █████╗ ██████╗  ██████╗ ██████╗ 
  ██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔═══██╗██╔══██╗
  ██████╔╝███████║███████╗   ██║   ██████╔╝█████╗  ███████║██║  ██║██║   ██║██████╔╝
  ██╔══██╗██╔══██║╚════██║   ██║   ██╔══██╗██╔══╝  ██╔══██║██║  ██║██║   ██║██╔══██╗
  ██║  ██║██║  ██║███████║   ██║   ██║  ██║███████╗██║  ██║██████╔╝╚██████╔╝██║  ██║
  ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝
              [ IP Multilateration & VPN De-Anonymization Engine ]
`

func main() {
	ipFlag := flag.String("ip", "", "Target IP address to locate and inspect")
	portsFlag := flag.String("ports", "80,443,22,53,8080", "Comma-separated TCP ports for RTT probing")
	samplesFlag := flag.Int("samples", 5, "Number of probe samples per landmark")
	l7URLFlag := flag.String("l7-url", "", "Optional HTTP/HTTPS endpoint on target for L4 vs L7 timing differential")
	outJSONFlag := flag.String("out", "", "Output JSON report path")
	mapHTMLFlag := flag.String("map", "map_result.html", "Output interactive HTML Leaflet map path")
	jsonOnlyFlag := flag.Bool("json", false, "Output strictly raw JSON to stdout")

	flag.Parse()

	if *ipFlag == "" {
		if !*jsonOnlyFlag {
			fmt.Print(banner)
		}
		fmt.Println("Uso: rastreador -ip <TARGET_IP> [opciones]")
		fmt.Println("\nOpciones:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !*jsonOnlyFlag {
		fmt.Print(banner)
		fmt.Printf("[*] Iniciando rastreo y multilateración sobre objetivo: %s\n\n", *ipFlag)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// 1. Reconnaissance & BGP Classification
	if !*jsonOnlyFlag {
		fmt.Println("[+] Paso 1/3: Reconocimiento de BGP, ASN e infraestructura...")
	}
	reconInfo, err := recon.QueryIP(ctx, *ipFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Error en reconocimiento: %v\n", err)
		os.Exit(1)
	}

	if !*jsonOnlyFlag {
		fmt.Printf("    • ASN: AS%d (%s)\n", reconInfo.ASN, reconInfo.ASOrg)
		fmt.Printf("    • País/ISP: %s / %s\n", reconInfo.Country, reconInfo.ISP)
		if reconInfo.Hostname != "" {
			fmt.Printf("    • Hostname PTR: %s\n", reconInfo.Hostname)
		}
		if reconInfo.AirportCode != "" {
			fmt.Printf("    • IATA Metro Code detectado: %s\n", reconInfo.AirportCode)
		}
		fmt.Printf("    • Clasificación de Red: [%s]\n", reconInfo.Confidence)
	}

	// 2. Multilateration Probing (CBG)
	if !*jsonOnlyFlag {
		fmt.Println("\n[+] Paso 2/3: Ejecutando sondas de retardo (RTT) desde red de landmarks...")
	}

	ports := parsePorts(*portsFlag)
	activeLandmarks := multilat.MultiProbeExecution(ctx, *ipFlag, multilat.DefaultVantageLandmarks, *samplesFlag)

	// If local active probes have high latency or are limited, calibrate with BGP fallback landmarks
	if len(activeLandmarks) == 0 {
		// Fallback calibration probe
		minRTT, _, probeErr := multilat.ProbeTarget(ctx, *ipFlag, ports, *samplesFlag)
		if probeErr == nil && minRTT > 0 {
			activeLandmarks = append(activeLandmarks, multilat.Landmark{
				ID:        "LOCAL-PROBE",
				Name:      "Local Vantage Node",
				City:      "Local",
				Country:   reconInfo.Country,
				Location:  multilat.Point{Lat: reconInfo.Latitude, Lon: reconInfo.Longitude},
				MinRTT:    minRTT,
				MaxRadius: multilat.ConstraintRadiusFromRTT(minRTT),
				Samples:   *samplesFlag,
			})
		}
	}

	// Solve coordinates
	var multiResult multilat.MultilaterationResult
	if len(activeLandmarks) > 0 {
		multiResult = multilat.SolveCentroidLeastSquares(activeLandmarks)
	} else if reconInfo.Latitude != 0 || reconInfo.Longitude != 0 {
		multiResult = multilat.MultilaterationResult{
			EstimatedPoint: multilat.Point{Lat: reconInfo.Latitude, Lon: reconInfo.Longitude},
			ConfidenceKm:   50.0,
			UsedLandmarks:  activeLandmarks,
		}
	}

	if !*jsonOnlyFlag {
		fmt.Printf("    • Sondas con respuesta: %d landmarks\n", len(activeLandmarks))
		for _, lm := range activeLandmarks {
			fmt.Printf("      - %s (%s): RTT Mínimo = %.2f ms | Radio Máximo CBG = %.0f km\n",
				lm.Name, lm.City, lm.MinRTT, lm.MaxRadius)
		}
		fmt.Printf("    • Coordenadas Físicas Estimadas: %.4f, %.4f (Radio de Confianza: ±%.1f km)\n",
			multiResult.EstimatedPoint.Lat, multiResult.EstimatedPoint.Lon, multiResult.ConfidenceKm)
	}

	// 3. Tunnel & Encapsulation Timing Differential
	if !*jsonOnlyFlag {
		fmt.Println("\n[+] Paso 3/3: Análisis de Túneles, MTU y diferencial L4/L7...")
	}
	tunnelResult, _ := tunnel.AnalyzeTunnel(ctx, *ipFlag, *l7URLFlag)
	if !*jsonOnlyFlag && tunnelResult != nil {
		for _, finding := range tunnelResult.Findings {
			fmt.Printf("    • %s\n", finding)
		}
	}

	// Build Full Report
	fullReport := &report.FullScanReport{
		TargetIP:        *ipFlag,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		ReconInfo:       reconInfo,
		Multilateration: &multiResult,
		TunnelAnalysis:  tunnelResult,
	}

	// Generate Map
	if *mapHTMLFlag != "" && multiResult.EstimatedPoint.Lat != 0 {
		if err := report.GenerateHTMLMap(fullReport, *mapHTMLFlag); err == nil {
			if !*jsonOnlyFlag {
				fmt.Printf("\n[✓] Mapa interactivo generado en: %s\n", *mapHTMLFlag)
			}
		}
	}

	// Output JSON if requested
	if *outJSONFlag != "" {
		data, _ := json.MarshalIndent(fullReport, "", "  ")
		os.WriteFile(*outJSONFlag, data, 0644)
		if !*jsonOnlyFlag {
			fmt.Printf("[✓] Reporte JSON exportado a: %s\n", *outJSONFlag)
		}
	}

	if *jsonOnlyFlag {
		data, _ := json.MarshalIndent(fullReport, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("\n[✓] Rastreo completado con éxito.")
	}
}

func parsePorts(portStr string) []int {
	parts := strings.Split(portStr, ",")
	ports := make([]int, 0, len(parts))
	for _, p := range parts {
		if val, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && val > 0 && val <= 65535 {
			ports = append(ports, val)
		}
	}
	return ports
}
