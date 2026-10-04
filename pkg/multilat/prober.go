package multilat

import (
	"context"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"time"
)

// Known IXP and Metro coordinate database for backbone hops
var knownMetroHubs = map[string]Point{
	"MAD": {Lat: 40.4168, Lon: -3.7038},  // Madrid / ESPANIX
	"BCN": {Lat: 41.3851, Lon: 2.1734},   // Barcelona / CATNIX
	"LIS": {Lat: 38.7223, Lon: -9.1393},  // Lisbon / GigaPIX
	"FRA": {Lat: 50.1109, Lon: 8.6821},   // Frankfurt / DE-CIX
	"AMS": {Lat: 52.3676, Lon: 4.9041},   // Amsterdam / AMS-IX
	"LON": {Lat: 51.5074, Lon: -0.1278},  // London / LINX
	"PAR": {Lat: 48.8566, Lon: 2.3522},   // Paris / France-IX
	"MIL": {Lat: 45.4642, Lon: 9.1900},   // Milan / MIX
	"ZRH": {Lat: 47.3769, Lon: 8.5417},   // Zurich / SwissIX
	"IAD": {Lat: 38.9531, Lon: -77.4565}, // Washington DC / Equinix Ashburn
	"JFK": {Lat: 40.6413, Lon: -73.7781}, // New York
	"ORD": {Lat: 41.8781, Lon: -87.6298}, // Chicago
	"DFW": {Lat: 32.7767, Lon: -96.7970}, // Dallas
	"SFO": {Lat: 37.7749, Lon: -122.4194},// San Francisco
	"LAX": {Lat: 34.0522, Lon: -118.2437},// Los Angeles
	"MIA": {Lat: 25.7617, Lon: -80.1918}, // Miami / NOTA
	"TYO": {Lat: 35.6762, Lon: 139.6503}, // Tokyo / JPIX
	"SIN": {Lat: 1.3521, Lon: 103.8198},  // Singapore / SGIX
	"SYD": {Lat: -33.8688, Lon: 151.2093},// Sydney
	"GRU": {Lat: -23.5505, Lon: -46.6333},// Sao Paulo / PTT.br
}

// LocalVantagePoint default coordinates (Spain / Madrid hub)
var DefaultLocalVantage = Point{Lat: 40.4168, Lon: -3.7038}

// ProbeTarget measures the RTT between local vantage point and target
func ProbeTarget(ctx context.Context, ip string, ports []int, samples int) (float64, int, error) {
	if samples <= 0 {
		samples = 5
	}
	if len(ports) == 0 {
		ports = []int{80, 443, 53, 22, 8080}
	}

	var bestPort int
	for _, port := range ports {
		addr := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err == nil {
			conn.Close()
			bestPort = port
			break
		}
	}

	if bestPort == 0 {
		bestPort = ports[0]
	}

	addr := fmt.Sprintf("%s:%d", ip, bestPort)
	rtts := make([]float64, 0, samples)

	for i := 0; i < samples; i++ {
		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		default:
		}

		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		dur := time.Since(start)

		if conn != nil {
			conn.Close()
		}

		if err == nil || isConnectionRefused(err) {
			rttMs := float64(dur.Microseconds()) / 1000.0
			rtts = append(rtts, rttMs)
		}

		time.Sleep(20 * time.Millisecond)
	}

	if len(rtts) == 0 {
		return 0, 0, fmt.Errorf("no response from %s", ip)
	}

	sort.Float64s(rtts)
	return rtts[0], bestPort, nil
}

// PerformMultiVantageProbing performs active local CBG constraint + route hop discovery
func PerformMultiVantageProbing(ctx context.Context, ip string, ports []int, samples int, registryLat, registryLon float64, metroCode string, isAnycast bool) ([]Landmark, Point, float64) {
	landmarks := make([]Landmark, 0)

	// 1. Measure direct RTT from local operator vantage point
	minRTT, _, err := ProbeTarget(ctx, ip, ports, samples)
	var localRadius float64
	if err == nil && minRTT > 0 {
		localRadius = ConstraintRadiusFromRTT(minRTT)
		landmarks = append(landmarks, Landmark{
			ID:        "LOCAL-VANTAGE",
			Name:      "Operador Local (Vantage Node)",
			City:      "Madrid Hub",
			Country:   "ES",
			Location:  DefaultLocalVantage,
			MinRTT:    minRTT,
			MaxRadius: localRadius,
			Samples:   samples,
			Type:      "local_vantage",
		})
	}

	// 2. Discover intermediate route landmarks (e.g. Metro POP / Airport code in PTR)
	if metroCode != "" {
		if pt, exists := knownMetroHubs[strings.ToUpper(metroCode)]; exists {
			landmarks = append(landmarks, Landmark{
				ID:        fmt.Sprintf("POP-%s", metroCode),
				Name:      fmt.Sprintf("Metro POP Gateway (%s)", metroCode),
				City:      metroCode,
				Country:   "",
				Location:  pt,
				MinRTT:    math.Max(2.0, minRTT*0.5),
				MaxRadius: ConstraintRadiusFromRTT(math.Max(2.0, minRTT*0.5)),
				Samples:   samples,
				Type:      "hop_landmark",
			})
		}
	}

	// 3. Anchor with BGP Registry / Geolocation if available
	hasRegistryCoords := (registryLat != 0 || registryLon != 0)
	if hasRegistryCoords {
		targetPt := Point{Lat: registryLat, Lon: registryLon}
		distToLocal := DistanceHaversine(DefaultLocalVantage, targetPt)

		// Check speed of light in fiber constraint:
		// If local ping is e.g. 7ms (radius 700km) but registry says USA (6000km), it's Anycast edge POP in Europe!
		if localRadius > 0 && distToLocal > localRadius*1.3 {
			// Speed of light physical violation -> Anycast Edge Node
			landmarks = append(landmarks, Landmark{
				ID:        "BGP-ORIGIN",
				Name:      "BGP Origin Registry (Anycast Parent)",
				City:      "Datacenter",
				Country:   "",
				Location:  targetPt,
				MinRTT:    minRTT,
				MaxRadius: math.Max(100.0, distToLocal),
				Samples:   samples,
				Type:      "bgp_anchor",
			})

			// Target physical active server is the local edge POP serving the request
			estPoint := DefaultLocalVantage
			confidence := math.Max(20.0, localRadius*0.5)
			return landmarks, estPoint, confidence
		}

		// Valid physical target matching speed of light
		landmarks = append(landmarks, Landmark{
			ID:        "TARGET-LOCATION",
			Name:      "Target Host Location",
			City:      "Target",
			Country:   "",
			Location:  targetPt,
			MinRTT:    minRTT,
			MaxRadius: math.Max(25.0, minRTT*5.0),
			Samples:   samples,
			Type:      "target_location",
		})

		return landmarks, targetPt, math.Max(15.0, minRTT*4.0)
	}

	// 4. Fallback solver if no registry coordinates were available
	if len(landmarks) > 0 {
		res := SolveCentroidLeastSquares(landmarks)
		return landmarks, res.EstimatedPoint, res.ConfidenceKm
	}

	return landmarks, DefaultLocalVantage, 100.0
}

func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "refused") || strings.Contains(s, "reset")
}
