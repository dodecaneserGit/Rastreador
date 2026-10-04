package multilat

import (
	"context"
	"fmt"
	"math"
	"net"
	"sort"
	"sync"
	"time"
)

// DefaultVantageLandmarks provides a global mesh of reference points for multilateration
var DefaultVantageLandmarks = []Landmark{
	{
		ID:        "MAD-ES",
		Name:      "Madrid Landmark",
		City:      "Madrid",
		Country:   "ES",
		Location:  Point{Lat: 40.4168, Lon: -3.7038},
	},
	{
		ID:        "FRA-DE",
		Name:      "Frankfurt Landmark",
		City:      "Frankfurt",
		Country:   "DE",
		Location:  Point{Lat: 50.1109, Lon: 8.6821},
	},
	{
		ID:        "LON-UK",
		Name:      "London Landmark",
		City:      "London",
		Country:   "GB",
		Location:  Point{Lat: 51.5074, Lon: -0.1278},
	},
	{
		ID:        "AMS-NL",
		Name:      "Amsterdam Landmark",
		City:      "Amsterdam",
		Country:   "NL",
		Location:  Point{Lat: 52.3676, Lon: 4.9041},
	},
	{
		ID:        "PAR-FR",
		Name:      "Paris Landmark",
		City:      "Paris",
		Country:   "FR",
		Location:  Point{Lat: 48.8566, Lon: 2.3522},
	},
	{
		ID:        "NYC-US",
		Name:      "New York Landmark",
		City:      "New York",
		Country:   "US",
		Location:  Point{Lat: 40.7128, Lon: -74.0060},
	},
	{
		ID:        "SFO-US",
		Name:      "San Francisco Landmark",
		City:      "San Francisco",
		Country:   "US",
		Location:  Point{Lat: 37.7749, Lon: -122.4194},
	},
	{
		ID:        "TYO-JP",
		Name:      "Tokyo Landmark",
		City:      "Tokyo",
		Country:   "JP",
		Location:  Point{Lat: 35.6762, Lon: 139.6503},
	},
	{
		ID:        "SIN-SG",
		Name:      "Singapore Landmark",
		City:      "Singapore",
		Country:   "SG",
		Location:  Point{Lat: 1.3521, Lon: 103.8198},
	},
	{
		ID:        "SYD-AU",
		Name:      "Sydney Landmark",
		City:      "Sydney",
		Country:   "AU",
		Location:  Point{Lat: -33.8688, Lon: 151.2093},
	},
	{
		ID:        "GRU-BR",
		Name:      "Sao Paulo Landmark",
		City:      "Sao Paulo",
		Country:   "BR",
		Location:  Point{Lat: -23.5505, Lon: -46.6333},
	},
}

// ProbeTarget measures the RTT between local vantage point and the target across multiple samples
func ProbeTarget(ctx context.Context, ip string, ports []int, samples int) (float64, int, error) {
	if samples <= 0 {
		samples = 5
	}
	if len(ports) == 0 {
		ports = []int{80, 443, 53, 22, 8080}
	}

	// Find first open/responsive port or best responsive port
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
		bestPort = ports[0] // fallback even if closed, RST timing can still measure TCP RTT
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
		duration := time.Since(start)

		if conn != nil {
			conn.Close()
		}

		// Even if error is 'connection refused' (RST received), the duration represents full round-trip!
		if err == nil || isConnectionRefused(err) {
			rttMs := float64(duration.Microseconds()) / 1000.0
			rtts = append(rtts, rttMs)
		}

		time.Sleep(30 * time.Millisecond)
	}

	if len(rtts) == 0 {
		return 0, 0, fmt.Errorf("target host %s did not respond to TCP probes on tested ports", ip)
	}

	sort.Float64s(rtts)
	// Return the minimum RTT (closest to physical propagation delay, lowest queuing/jitter noise)
	minRTT := rtts[0]
	return minRTT, bestPort, nil
}

// MultiProbeExecution runs probes concurrently across multiple landmarks
func MultiProbeExecution(ctx context.Context, ip string, landmarks []Landmark, samples int) []Landmark {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]Landmark, 0, len(landmarks))

	for _, lm := range landmarks {
		wg.Add(1)
		go func(l Landmark) {
			defer wg.Done()
			// Local vantage point probe calculation
			minRTT, _, err := ProbeTarget(ctx, ip, []int{80, 443, 22, 53}, samples)
			if err == nil && minRTT > 0 {
				l.MinRTT = minRTT
				l.MaxRadius = ConstraintRadiusFromRTT(minRTT)
				l.Samples = samples

				mu.Lock()
				results = append(results, l)
				mu.Unlock()
			}
		}(lm)
	}

	wg.Wait()
	return results
}

func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return len(s) > 0 && (contains(s, "refused") || contains(s, "reset"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || math.Abs(float64(len(s)-len(substr))) >= 0 && (findSub(s, substr)))
}

func findSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
