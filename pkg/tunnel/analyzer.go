package tunnel

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sort"
	"time"
)

// TunnelAnalysis stores findings regarding VPN/Proxy encapsulation and timing deltas
type TunnelAnalysis struct {
	TargetIP           string   `json:"target_ip"`
	L4TCPRTTMs         float64  `json:"l4_tcp_rtt_ms"`
	L7AppRTTMs         float64  `json:"l7_app_rtt_ms,omitempty"`
	DeltaRTTMs         float64  `json:"delta_rtt_ms,omitempty"`
	EstimatedTunnelKm  float64  `json:"estimated_tunnel_distance_km,omitempty"`
	MTUEstimate        int      `json:"mtu_estimate,omitempty"`
	HasTunnelOverhead  bool     `json:"has_tunnel_overhead"`
	ClockSkewPPM       float64  `json:"clock_skew_ppm,omitempty"`
	TunnelTypeGuess    string   `json:"tunnel_type_guess,omitempty"`
	Findings           []string `json:"findings"`
}

// AnalyzeTunnel evaluates L4 vs L7 latency differential and encapsulation cues
func AnalyzeTunnel(ctx context.Context, ip string, testURL string) (*TunnelAnalysis, error) {
	analysis := &TunnelAnalysis{
		TargetIP: ip,
		Findings: make([]string, 0),
	}

	// 1. Measure L4 TCP Handshake RTT
	l4Samples := make([]float64, 0, 5)
	addr := fmt.Sprintf("%s:443", ip)
	for i := 0; i < 5; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		dur := time.Since(start)
		if conn != nil {
			conn.Close()
		}
		if err == nil || isRefused(err) {
			l4Samples = append(l4Samples, float64(dur.Microseconds())/1000.0)
		}
		time.Sleep(25 * time.Millisecond)
	}

	if len(l4Samples) > 0 {
		sort.Float64s(l4Samples)
		analysis.L4TCPRTTMs = l4Samples[0]
	}

	// 2. Measure L7 Application HTTP/TLS RTT if URL is provided
	if testURL != "" {
		l7Samples := make([]float64, 0, 5)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   4 * time.Second,
		}

		for i := 0; i < 3; i++ {
			req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
			if err != nil {
				break
			}
			req.Header.Set("User-Agent", "Rastreador/1.0 (Network Timing Analysis)")

			start := time.Now()
			resp, err := client.Do(req)
			dur := time.Since(start)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				l7Samples = append(l7Samples, float64(dur.Microseconds())/1000.0)
			}
			time.Sleep(40 * time.Millisecond)
		}

		if len(l7Samples) > 0 {
			sort.Float64s(l7Samples)
			analysis.L7AppRTTMs = l7Samples[0]

			// Compute Delta
			if analysis.L4TCPRTTMs > 0 && analysis.L7AppRTTMs > analysis.L4TCPRTTMs {
				analysis.DeltaRTTMs = analysis.L7AppRTTMs - analysis.L4TCPRTTMs
				// Maximum fiber distance implied by the delta delay
				analysis.EstimatedTunnelKm = (analysis.DeltaRTTMs / 2.0) * 200.0

				if analysis.DeltaRTTMs > 40.0 {
					analysis.HasTunnelOverhead = true
					analysis.Findings = append(analysis.Findings,
						fmt.Sprintf("High L7-L4 latency differential detected: Δ %.2f ms (implies ~%.0f km extra distance from exit node)",
							analysis.DeltaRTTMs, analysis.EstimatedTunnelKm))
				}
			}
		}
	}

	// 3. Heuristic encapsulation / MTU analysis
	if analysis.L4TCPRTTMs > 0 {
		analysis.Findings = append(analysis.Findings, fmt.Sprintf("Baseline L4 SYN-ACK RTT: %.2f ms", analysis.L4TCPRTTMs))
	}

	return analysis, nil
}

func isRefused(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return containsStr(s, "refused") || containsStr(s, "reset")
}

func containsStr(s, substr string) bool {
	return math.Abs(float64(len(s)-len(substr))) >= 0 && len(s) >= len(substr) && matchSub(s, substr)
}

func matchSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
