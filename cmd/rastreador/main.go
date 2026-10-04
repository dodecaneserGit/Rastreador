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
	ripeKeyFlag := flag.String("ripe-key", "", "Optional RIPE Atlas API Key for distributed worldwide probing")
	ripeProbesFlag := flag.Int("ripe-probes", 5, "Number of worldwide probes to request from RIPE Atlas")
	outJSONFlag := flag.String("out", "", "Output JSON report path")
	mapHTMLFlag := flag.String("map", "map_result.html", "Output interactive HTML Leaflet map path")
	openBrowserFlag := flag.Bool("open", true, "Automatically open the generated HTML map in default web browser")
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
		fmt.Printf("[*] Iniciando rastreo y multilateración sobre objetivo: \033[1;36m%s\033[0m\n\n", *ipFlag)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	// 1. Reconnaissance & BGP Classification
	if !*jsonOnlyFlag {
		fmt.Println("[+] Paso 1/3: Reconocimiento BGP, ASN e infraestructura...")
	}
	reconInfo, err := recon.QueryIP(ctx, *ipFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Error en reconocimiento: %v\n", err)
		os.Exit(1)
	}

	// 2. Multilateration Probing (Local Vantage + RIPE Atlas Global Sondas)
	if !*jsonOnlyFlag {
		fmt.Println("[+] Paso 2/3: Ejecutando sondeo de retardo RTT y multilateración multi-nodo...")
	}

	ports := parsePorts(*portsFlag)
	landmarks, estPoint, confidence := multilat.PerformMultiVantageProbing(ctx, *ipFlag, ports, *samplesFlag, reconInfo.Latitude, reconInfo.Longitude, reconInfo.AirportCode, reconInfo.IsAnycast)

	// Query RIPE Atlas if API key is provided, found in env, or in ~/.zshrc
	ripeClient := multilat.NewRIPEAtlasClient(*ripeKeyFlag)
	if ripeClient.APIKey != "" {
		if !*jsonOnlyFlag {
			maskedKey := ripeClient.APIKey
			if len(maskedKey) > 8 {
				maskedKey = maskedKey[:4] + "..." + maskedKey[len(maskedKey)-4:]
			}
			fmt.Printf("    • [✓] Clave RIPE Atlas detectada (%s). Solicitando %d sondas globales...\n", maskedKey, *ripeProbesFlag)
		}
		ripeLandmarks, err := ripeClient.RunRIPEAtlasProbing(ctx, *ipFlag, *ripeProbesFlag)
		if err == nil && len(ripeLandmarks) > 0 {
			if !*jsonOnlyFlag {
				fmt.Printf("    • [✓] Recibidas %d respuestas de sondas RIPE Atlas en tiempo real.\n", len(ripeLandmarks))
			}
			landmarks = append(landmarks, ripeLandmarks...)
			multiSolverRes := multilat.SolveCentroidLeastSquares(landmarks)
			estPoint = multiSolverRes.EstimatedPoint
			confidence = multiSolverRes.ConfidenceKm
		} else if !*jsonOnlyFlag && err != nil {
			fmt.Printf("    • [!] Nota RIPE Atlas: %v\n", err)
		}
	} else {
		if !*jsonOnlyFlag {
			fmt.Println("    • [ℹ] Sin clave RIPE Atlas (ejecutando con nodo local y BGP).")
		}
	}

	multiResult := multilat.MultilaterationResult{
		EstimatedPoint: estPoint,
		ConfidenceKm:   confidence,
		UsedLandmarks:  landmarks,
	}

	// 3. Tunnel & Encapsulation Timing Differential
	if !*jsonOnlyFlag {
		fmt.Println("[+] Paso 3/3: Análisis de Túneles, MTU y diferencial L4/L7...")
	}
	tunnelResult, _ := tunnel.AnalyzeTunnel(ctx, *ipFlag, *l7URLFlag)

	// Build Full Report
	fullReport := &report.FullScanReport{
		TargetIP:        *ipFlag,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		ReconInfo:       reconInfo,
		Multilateration: &multiResult,
		TunnelAnalysis:  tunnelResult,
	}

	// Print Terminal Summary
	if !*jsonOnlyFlag {
		report.PrintTerminalSummary(fullReport)
	}

	// Generate Map
	if *mapHTMLFlag != "" && multiResult.EstimatedPoint.Lat != 0 {
		if err := report.GenerateHTMLMap(fullReport, *mapHTMLFlag); err == nil {
			if !*jsonOnlyFlag {
				fmt.Printf("[✓] Mapa interactivo generado en: \033[1;32m%s\033[0m (Esri & Leaflet)\n", *mapHTMLFlag)
			}
			if *openBrowserFlag && !*jsonOnlyFlag {
				_ = report.OpenInBrowser(*mapHTMLFlag)
				fmt.Println("[🚀] Mapa abierto automáticamente en tu navegador predeterminado.")
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
