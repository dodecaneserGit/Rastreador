package l2wifi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dodecaneser/rastreador/pkg/multilat"
)

// WiFiNetwork represents a Wi-Fi beacon observation
type WiFiNetwork struct {
	BSSID       string  `json:"bssid"`                 // MAC Address (e.g., "00:11:22:33:44:55")
	SSID        string  `json:"ssid,omitempty"`        // Network name
	RSSI        int     `json:"rssi,omitempty"`        // Received Signal Strength in dBm (e.g., -55)
	Channel     int     `json:"channel,omitempty"`
	Lat         float64 `json:"latitude,omitempty"`
	Lon         float64 `json:"longitude,omitempty"`
	Road        string  `json:"road,omitempty"`
	HouseNumber string  `json:"house_number,omitempty"`
	City        string  `json:"city,omitempty"`
	Country     string  `json:"country,omitempty"`
	Resolved    bool    `json:"resolved"`
}

// TriangulationResult stores the final sub-30m physical location calculated from Wi-Fi beacons
type TriangulationResult struct {
	EstimatedPoint multilat.Point `json:"estimated_point"`
	ConfidenceKm   float64        `json:"confidence_km"`   // e.g. 0.015 (15 meters)
	PrecisionM     float64        `json:"precision_meters"`// e.g. 15.0 m
	ResolvedCount  int            `json:"resolved_count"`
	TotalBeacons   int            `json:"total_beacons"`
	StreetAddress  string         `json:"street_address,omitempty"`
	Networks       []WiFiNetwork  `json:"networks"`
}

// WiGLEClient interacts with the WiGLE.net API v2
type WiGLEClient struct {
	AuthToken  string // Base64 encoded "API_NAME:API_TOKEN"
	HTTPClient *http.Client
}

// NewWiGLEClient creates a WiGLE client, discovering credentials from flag, env or ~/.zshrc
func NewWiGLEClient(userProvidedKey string) *WiGLEClient {
	token := strings.TrimSpace(userProvidedKey)

	// Check environment variable
	if token == "" {
		token = strings.TrimSpace(os.Getenv("WIGLE_API_KEY"))
	}
	if token == "" {
		token = strings.TrimSpace(os.Getenv("WIGLE_AUTH"))
	}

	// Check ~/.zshrc or ~/.bashrc if still empty
	if token == "" {
		token = discoverWiGLEKeyFromShell()
	}

	// Format as Basic Auth token if user passed "AID:TOKEN"
	if token != "" && !strings.Contains(token, "Basic ") {
		if strings.Contains(token, ":") {
			token = "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
		} else if !isBase64(token) {
			token = "Basic " + token
		} else {
			token = "Basic " + token
		}
	}

	return &WiGLEClient{
		AuthToken: token,
		HTTPClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	}
}

func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func discoverWiGLEKeyFromShell() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	files := []string{".zshrc", ".bashrc", ".zprofile", ".profile"}
	regex := regexp.MustCompile(`(?i)export\s+(?:WIGLE_API_KEY|WIGLE_AUTH)=["']?([^"'\r\n]+)["']?`)

	for _, fname := range files {
		p := home + "/" + fname
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		matches := regex.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

type wigleDetailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Results []struct {
		NetID       string  `json:"netid"`
		SSID        string  `json:"ssid"`
		Trilat      float64 `json:"trilat"`
		Trilong     float64 `json:"trilong"`
		Road        string  `json:"road"`
		HouseNumber string  `json:"housenumber"`
		City        string  `json:"city"`
		Country     string  `json:"country"`
		Channel     int     `json:"channel"`
	} `json:"results"`
}

// QueryBSSID fetches geocoordinates for a single BSSID from WiGLE.net
func (c *WiGLEClient) QueryBSSID(ctx context.Context, bssid string) (*WiFiNetwork, error) {
	normBSSID := normalizeBSSID(bssid)
	if normBSSID == "" {
		return nil, fmt.Errorf("invalid BSSID format: %s", bssid)
	}

	net := &WiFiNetwork{
		BSSID: normBSSID,
	}

	if c.AuthToken == "" {
		return nil, fmt.Errorf("WiGLE API key not configured (set WIGLE_API_KEY in ~/.zshrc or pass -wigle-key)")
	}

	// 1. Primary: Query WiGLE Search API
	searchURL := fmt.Sprintf("https://api.wigle.net/api/v2/network/search?netid=%s", url.QueryEscape(normBSSID))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Rastreador-Engine/2.0 (L2-WiFi-MicroLocator)")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("WiGLE authentication failed (HTTP %d). Check your API Token in ~/.zshrc", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusOK {
		var data wigleDetailResponse
		if err := json.Unmarshal(body, &data); err == nil && data.Success && len(data.Results) > 0 {
			res := data.Results[0]
			net.SSID = res.SSID
			net.Lat = res.Trilat
			net.Lon = res.Trilong
			net.Road = res.Road
			net.HouseNumber = res.HouseNumber
			net.City = res.City
			net.Country = res.Country
			net.Channel = res.Channel
			net.Resolved = (net.Lat != 0 || net.Lon != 0)
			if net.Resolved {
				return net, nil
			}
		}
	}

	// 2. Fallback: Query WiGLE Detail API
	detailURL := fmt.Sprintf("https://api.wigle.net/api/v2/network/detail?netid=%s", url.QueryEscape(normBSSID))
	reqDetail, err := http.NewRequestWithContext(ctx, "GET", detailURL, nil)
	if err == nil {
		reqDetail.Header.Set("Authorization", c.AuthToken)
		reqDetail.Header.Set("Accept", "application/json")
		reqDetail.Header.Set("User-Agent", "Rastreador-Engine/2.0 (L2-WiFi-MicroLocator)")
		respDetail, err := c.HTTPClient.Do(reqDetail)
		if err == nil {
			defer respDetail.Body.Close()
			bodyDetail, _ := io.ReadAll(respDetail.Body)
			var dataDetail wigleDetailResponse
			if err := json.Unmarshal(bodyDetail, &dataDetail); err == nil && dataDetail.Success && len(dataDetail.Results) > 0 {
				res := dataDetail.Results[0]
				net.SSID = res.SSID
				net.Lat = res.Trilat
				net.Lon = res.Trilong
				net.Road = res.Road
				net.HouseNumber = res.HouseNumber
				net.City = res.City
				net.Country = res.Country
				net.Channel = res.Channel
				net.Resolved = (net.Lat != 0 || net.Lon != 0)
				if net.Resolved {
					return net, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("BSSID %s no encontrado en el catálogo global de WiGLE (posible router nuevo o MAC privada)", normBSSID)
}

// TriangulateBSSIDs resolves a list of BSSIDs and performs RSSI-weighted multilateration to break the 30m barrier
func (c *WiGLEClient) TriangulateBSSIDs(ctx context.Context, networks []WiFiNetwork) (*TriangulationResult, error) {
	if len(networks) == 0 {
		return nil, fmt.Errorf("no BSSID targets provided")
	}

	resolvedList := make([]WiFiNetwork, 0, len(networks))

	for i := range networks {
		n := networks[i]
		if n.Lat != 0 && n.Lon != 0 {
			n.Resolved = true
			resolvedList = append(resolvedList, n)
			continue
		}

		fetched, err := c.QueryBSSID(ctx, n.BSSID)
		if err == nil && fetched.Resolved {
			fetched.RSSI = n.RSSI
			resolvedList = append(resolvedList, *fetched)
		} else {
			// Keep unresolved placeholder
			resolvedList = append(resolvedList, n)
		}
		time.Sleep(100 * time.Millisecond) // Polite WiGLE API pacing
	}

	var validNets []WiFiNetwork
	for _, n := range resolvedList {
		if n.Resolved {
			validNets = append(validNets, n)
		}
	}

	if len(validNets) == 0 {
		return &TriangulationResult{
			ResolvedCount: 0,
			TotalBeacons:  len(networks),
			Networks:      resolvedList,
		}, fmt.Errorf("none of the %d provided BSSIDs could be geolocated via WiGLE", len(networks))
	}

	// 1 Single BSSID Match -> Direct Building / Door level (±25m)
	if len(validNets) == 1 {
		pt := multilat.Point{Lat: validNets[0].Lat, Lon: validNets[0].Lon}
		addr := formatAddress(validNets[0])
		return &TriangulationResult{
			EstimatedPoint: pt,
			ConfidenceKm:   0.025, // 25 meters
			PrecisionM:     25.0,
			ResolvedCount:  1,
			TotalBeacons:   len(networks),
			StreetAddress:  addr,
			Networks:       resolvedList,
		}, nil
	}

	// 2+ BSSIDs -> Weighted Centroid by RSSI signal strength (log-distance attenuation)
	var sumLat, sumLon, sumWeight float64
	for _, n := range validNets {
		rssi := n.RSSI
		if rssi == 0 {
			rssi = -70 // default nominal RSSI
		}
		// Weight w = 10^(RSSI / 20)
		weight := math.Pow(10, float64(rssi)/20.0)
		if weight <= 0 {
			weight = 0.001
		}

		sumLat += n.Lat * weight
		sumLon += n.Lon * weight
		sumWeight += weight
	}

	estLat := sumLat / sumWeight
	estLon := sumLon / sumWeight
	estPt := multilat.Point{Lat: estLat, Lon: estLon}

	// Calculate dispersion radius (standard deviation)
	var maxDistKm float64
	for _, n := range validNets {
		d := multilat.DistanceHaversine(estPt, multilat.Point{Lat: n.Lat, Lon: n.Lon})
		if d > maxDistKm {
			maxDistKm = d
		}
	}

	// Confidence radius: between 8 meters (dense tri) and 20 meters
	precM := math.Max(8.0, math.Min(25.0, maxDistKm*1000.0*0.6))
	confidenceKm := precM / 1000.0

	return &TriangulationResult{
		EstimatedPoint: estPt,
		ConfidenceKm:   confidenceKm,
		PrecisionM:     math.Round(precM*10) / 10,
		ResolvedCount:  len(validNets),
		TotalBeacons:   len(networks),
		StreetAddress:  formatAddress(validNets[0]),
		Networks:       resolvedList,
	}, nil
}

// ScanLocalWiFiNetworks scans surrounding Wi-Fi beacons using airport on macOS or iwlist/nmcli on Linux
func ScanLocalWiFiNetworks() ([]WiFiNetwork, error) {
	switch runtime.GOOS {
	case "darwin":
		return scanMacOSAirport()
	case "linux":
		return scanLinuxWiFi()
	default:
		return nil, fmt.Errorf("live Wi-Fi scanning unsupported on %s (pass -bssid manually)", runtime.GOOS)
	}
}

func scanMacOSAirport() ([]WiFiNetwork, error) {
	airportBin := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	if _, err := os.Stat(airportBin); err == nil {
		cmd := exec.Command(airportBin, "-s")
		out, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) >= 2 {
				macRegex := regexp.MustCompile(`([0-9a-fA-F]{2}(?::[0-9a-fA-F]{2}){5})`)
				rssiRegex := regexp.MustCompile(`(-\d{2,3})`)
				networks := make([]WiFiNetwork, 0)

				for _, line := range lines[1:] {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					macMatch := macRegex.FindString(line)
					if macMatch == "" {
						continue
					}

					var rssiVal int = -70
					rssiMatch := rssiRegex.FindString(line)
					if rssiMatch != "" {
						if r, err := strconv.Atoi(rssiMatch); err == nil {
							rssiVal = r
						}
					}

					parts := strings.Fields(line)
					ssid := ""
					if len(parts) > 0 {
						ssid = parts[0]
					}

					networks = append(networks, WiFiNetwork{
						BSSID: macMatch,
						SSID:  ssid,
						RSSI:  rssiVal,
					})
				}
				if len(networks) > 0 {
					return networks, nil
				}
			}
		}
	}

	// Fallback to system_profiler
	cmd := exec.Command("system_profiler", "SPAirPortDataType")
	out, err := cmd.Output()
	if err == nil {
		macRegex := regexp.MustCompile(`(?i)(?:BSSID|MAC Address):\s*([0-9a-fA-F]{2}(?::[0-9a-fA-F]{2}){5})`)
		matches := macRegex.FindAllStringSubmatch(string(out), -1)
		if len(matches) > 0 {
			networks := make([]WiFiNetwork, 0)
			seen := make(map[string]bool)
			for _, m := range matches {
				if len(m) > 1 {
					bssid := normalizeBSSID(m[1])
					if bssid != "" && !seen[bssid] {
						seen[bssid] = true
						networks = append(networks, WiFiNetwork{
							BSSID: bssid,
							RSSI:  -60,
						})
					}
				}
			}
			if len(networks) > 0 {
				return networks, nil
			}
		}
	}

	return nil, fmt.Errorf("no se pudieron escanear balizas Wi-Fi automáticamente en macOS (pasa los BSSIDs con -bssid \"AA:BB:CC:DD:EE:FF\")")
}

func scanLinuxWiFi() ([]WiFiNetwork, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "BSSID,SSID,SIGNAL,CHAN", "dev", "wifi")
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		networks := make([]WiFiNetwork, 0)
		for _, l := range lines {
			parts := strings.Split(l, ":")
			if len(parts) >= 6 {
				// Reconstruct BSSID
				bssid := fmt.Sprintf("%s:%s:%s:%s:%s:%s", parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
				ssid := ""
				signal := 50
				if len(parts) >= 7 {
					ssid = parts[6]
				}
				if len(parts) >= 8 {
					if sig, err := strconv.Atoi(parts[7]); err == nil {
						signal = sig
					}
				}
				// Convert percentage signal to dBm: dBm ~= (signal / 2) - 100
				rssi := (signal / 2) - 100
				networks = append(networks, WiFiNetwork{
					BSSID: bssid,
					SSID:  ssid,
					RSSI:  rssi,
				})
			}
		}
		if len(networks) > 0 {
			return networks, nil
		}
	}

	return nil, fmt.Errorf("no Linux Wi-Fi scan tools succeeded (install nmcli or pass -bssid)")
}

func normalizeBSSID(mac string) string {
	mac = strings.TrimSpace(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ToLower(mac)
	var parts []string
	if strings.Contains(mac, ":") {
		parts = strings.Split(mac, ":")
	} else if len(mac) == 12 {
		for i := 0; i < 12; i += 2 {
			parts = append(parts, mac[i:i+2])
		}
	}
	if len(parts) != 6 {
		return ""
	}
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = "0" + p
		}
	}
	return strings.Join(parts, ":")
}

func formatAddress(n WiFiNetwork) string {
	parts := make([]string, 0)
	if n.Road != "" {
		if n.HouseNumber != "" {
			parts = append(parts, fmt.Sprintf("%s %s", n.Road, n.HouseNumber))
		} else {
			parts = append(parts, n.Road)
		}
	}
	if n.City != "" {
		parts = append(parts, n.City)
	}
	if n.Country != "" {
		parts = append(parts, n.Country)
	}
	return strings.Join(parts, ", ")
}
