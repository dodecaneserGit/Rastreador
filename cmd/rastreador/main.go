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

	"github.com/dodecaneser/rastreador/pkg/l2wifi"
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
              [ IP Multilateration & L2 Wi-Fi De-Anonymization Engine ]
`

func main() {
	ipFlag := flag.String("ip", "", "Target IP address to locate and inspect")
	bssidFlag := flag.String("bssid", "", "Comma-separated target Wi-Fi BSSID MAC addresses for L2 micro-triangulation (<30m)")
	scanWiFiFlag := flag.Bool("scan-wifi", false, "Scan local surrounding Wi-Fi beacons and triangulate physical location")
	wigleKeyFlag := flag.String("wigle-key", "", "Optional WiGLE API Key (API_NAME:API_TOKEN or Base64) for BSSID resolution")
	portsFlag := flag.String("ports", "80,443,22,53,8080", "Comma-separated TCP ports for RTT probing")
	samplesFlag := flag.Int("samples", 5, "Number of probe samples per landmark")
	l7URLFlag := flag.String("l7-url", "", "Optional HTTP/HTTPS endpoint on target for L4 vs L7 timing differential")
	ripeKeyFlag := flag.String("ripe-key", "", "Optional RIPE Atlas API Key for distributed worldwide probing")
	ripeProbesFlag := flag.Int("ripe-probes", 4, "Number of worldwide probes to request from RIPE Atlas")
	outJSONFlag := flag.String("out", "", "Output JSON report path")
	mapHTMLFlag := flag.String("map", "map_result.html", "Output interactive HTML Leaflet map path")
	openBrowserFlag := flag.Bool("open", true, "Automatically open the generated HTML map in default web browser")
	jsonOnlyFlag := flag.Bool("json", false, "Output strictly raw JSON to stdout")

	flag.Parse()

	if *ipFlag == "" && *bssidFlag == "" && !*scanWiFiFlag {
		if !*jsonOnlyFlag {
			fmt.Print(banner)
		}
		fmt.Println("Uso: rastreador -ip <TARGET_IP> [opciones]")
		fmt.Println("     rastreador -bssid <MAC_1,MAC_2,...> [opciones]")
		fmt.Println("     rastreador -scan-wifi [opciones]")
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

	var reconInfo *recon.IPInfo
	var landmarks []multilat.Landmark
	var estPoint multilat.Point
	var confidence float64 = 25.0

	// 1. Reconnaissance & BGP Classification (if IP is provided)
	if *ipFlag != "" {
		if !*jsonOnlyFlag {
			fmt.Println("[+] Paso 1/3: Reconocimiento BGP, ASN e infraestructura...")
		}
		var err error
		reconInfo, err = recon.QueryIP(ctx, *ipFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Error en reconocimiento: %v\n", err)
			os.Exit(1)
		}

		// 2. Multilateration Probing (Local Vantage + RIPE Atlas Global Sondas)
		if !*jsonOnlyFlag {
			fmt.Println("[+] Paso 2/3: Ejecutando sondeo de retardo RTT y multilateración multi-nodo...")
		}

		ports := parsePorts(*portsFlag)
		landmarks, estPoint, confidence = multilat.PerformMultiVantageProbing(ctx, *ipFlag, ports, *samplesFlag, reconInfo.Latitude, reconInfo.Longitude, reconInfo.PrecisionKm, reconInfo.AirportCode, reconInfo.IsAnycast)

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

				// If no authoritative geolocation was available, solve via least-squares
				if reconInfo.Latitude == 0 && reconInfo.Longitude == 0 {
					multiSolverRes := multilat.SolveCentroidLeastSquares(landmarks)
					estPoint = multiSolverRes.EstimatedPoint
					confidence = multiSolverRes.ConfidenceKm
				}
			} else if !*jsonOnlyFlag && err != nil {
				fmt.Printf("    • [!] Nota RIPE Atlas: %v (usando sondas locales y BGP)\n", err)
			}
		} else {
			if !*jsonOnlyFlag {
				fmt.Println("    • [ℹ] Sin clave RIPE Atlas (ejecutando con nodo local y BGP).")
			}
		}
	} else {
		reconInfo = &recon.IPInfo{
			IP:         "N/A (L2 Wi-Fi Mode)",
			Confidence: "L2_WIFI_DIRECT",
		}
	}

	multiResult := multilat.MultilaterationResult{
		EstimatedPoint: estPoint,
		ConfidenceKm:   confidence,
		UsedLandmarks:  landmarks,
	}

	// 3. Tunnel & Encapsulation Timing Differential
	var tunnelResult *tunnel.TunnelAnalysis
	if *ipFlag != "" {
		if !*jsonOnlyFlag {
			fmt.Println("[+] Paso 3/3: Análisis de Túneles, MTU y diferencial L4/L7...")
		}
		tunnelResult, _ = tunnel.AnalyzeTunnel(ctx, *ipFlag, *l7URLFlag)
	}

	// 4. L2 Wi-Fi Micro-Triangulation (<30m) via WiGLE
	var wifiTriResult *l2wifi.TriangulationResult
	if *bssidFlag != "" || *scanWiFiFlag {
		if !*jsonOnlyFlag {
			fmt.Println("\n[+] Paso L2: Micro-Localización Wi-Fi por balizas BSSID (WiGLE Engine)...")
		}

		var targetNets []l2wifi.WiFiNetwork
		if *scanWiFiFlag {
			scanned, err := l2wifi.ScanLocalWiFiNetworks()
			if err == nil && len(scanned) > 0 {
				if !*jsonOnlyFlag {
					fmt.Printf("    • [✓] Escaneadas %d balizas Wi-Fi en el entorno físico local.\n", len(scanned))
				}
				targetNets = append(targetNets, scanned...)
			} else if !*jsonOnlyFlag {
				fmt.Printf("    • [!] Nota escaneo Wi-Fi local: %v\n", err)
			}
		}

		if *bssidFlag != "" {
			parts := strings.Split(*bssidFlag, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					targetNets = append(targetNets, l2wifi.WiFiNetwork{BSSID: p})
				}
			}
		}

		if len(targetNets) > 0 {
			wigleClient := l2wifi.NewWiGLEClient(*wigleKeyFlag)
			if wigleClient.AuthToken != "" {
				res, err := wigleClient.TriangulateBSSIDs(ctx, targetNets)
				if err == nil && res.ResolvedCount > 0 {
					wifiTriResult = res
					if !*jsonOnlyFlag {
						fmt.Printf("    • [✓] Triangulación WiGLE exitosa: Coordenadas %.6f, %.6f (Precisión: ±%.1fm)\n",
							res.EstimatedPoint.Lat, res.EstimatedPoint.Lon, res.PrecisionM)
					}
					// Update final estimated point with sub-30m Wi-Fi precision
					if res.EstimatedPoint.Lat != 0 {
						estPoint = res.EstimatedPoint
						confidence = res.ConfidenceKm
						multiResult.EstimatedPoint = estPoint
						multiResult.ConfidenceKm = confidence
					}
				} else if !*jsonOnlyFlag && err != nil {
					fmt.Printf("    • [!] Nota WiGLE: %v\n", err)
				}
			} else if !*jsonOnlyFlag {
				fmt.Println("    • [ℹ] Clave WiGLE no configurada (exporta WIGLE_API_KEY o pasa -wigle-key).")
			}
		}
	}

	// Build Full Report
	fullReport := &report.FullScanReport{
		TargetIP:          *ipFlag,
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		ReconInfo:         reconInfo,
		Multilateration:   &multiResult,
		TunnelAnalysis:    tunnelResult,
		WiFiTriangulation: wifiTriResult,
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
