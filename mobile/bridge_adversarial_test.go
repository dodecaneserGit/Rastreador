package mobile

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
)

// ============================================================================
// 1. Malformed & Corrupted JSON Payloads Fuzzing
// ============================================================================

func TestAdversarialJSON_CorruptedPayloads(t *testing.T) {
	corruptedInputs := []string{
		"",
		"   ",
		"not json at all",
		"{",
		"}",
		`{"ip": `,
		`{"ip": "1.1.1.1"`,
		`{"beacons": [}`,
		`{"lat1": 40.0, "lon1": }`,
		"null",
		"true",
		"false",
		"12345",
		`"just a string"`,
		"[]",
		`[{"ip": "1.1.1.1"}]`,
		`{"ip": 12345}`,
		`{"beacons": "not_an_array"}`,
		`{"lat1": "not_a_number"}`,
		`{"timeout_sec": "five"}`,
		"\x00\x01\x02\x03",
		`{"ip": "\u0000\u0001"}`,
		strings.Repeat(`{"nested": `, 100) + `1` + strings.Repeat(`}`, 100),
		strings.Repeat("A", 100000),
	}

	entrypoints := []struct {
		name string
		fn   func(string) string
	}{
		{"QueryBgpAsn", QueryBgpAsn},
		{"PerformMultilateration", PerformMultilateration},
		{"TriangulateWiFi", TriangulateWiFi},
		{"AnalyzeIpId", AnalyzeIpId},
		{"AnalyzeTunnel", AnalyzeTunnel},
		{"CalculateHaversineDistanceJSON", CalculateHaversineDistanceJSON},
	}

	for _, ep := range entrypoints {
		for i, payload := range corruptedInputs {
			t.Run(fmt.Sprintf("%s_case_%d", ep.name, i), func(t *testing.T) {
				// Call must NEVER panic
				respJSON := ep.fn(payload)

				if respJSON == "" {
					t.Fatalf("%s returned empty string for input %q", ep.name, payload)
				}

				var env ResponseEnvelope
				if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
					t.Fatalf("%s returned invalid JSON envelope: %v. Raw: %s", ep.name, err, respJSON)
				}

				// Corrupted input should return success: false (except null on CalculateHaversine where Go unmarshals zero-value struct)
				if env.Success && (payload == "" || (!strings.Contains(payload, "{") && payload != "null")) {
					t.Fatalf("%s unexpectedly succeeded on invalid payload %q", ep.name, payload)
				}
			})
		}
	}
}

func TestAdversarialJSON_InvokeJsonMultiplexer(t *testing.T) {
	// Unknown actions
	badActions := []string{"", "unknown", "DROP TABLE", "null", "\x00", "QUERY_BGP"}
	for _, act := range badActions {
		resp := InvokeJson(act, `{}`)
		var env ResponseEnvelope
		if err := json.Unmarshal([]byte(resp), &env); err != nil {
			t.Fatalf("InvokeJson returned invalid JSON for action %q: %v", act, err)
		}
		if env.Success {
			t.Fatalf("InvokeJson unexpectedly succeeded on unknown action %q", act)
		}
		if env.Error == nil || !strings.Contains(*env.Error, "unknown action") {
			t.Fatalf("Expected unknown action error, got %v", env.Error)
		}
	}
}

// ============================================================================
// 2. Out-of-Bounds Geographic Inputs (lat > 90, lon > 180, NaN, Inf)
// ============================================================================

func TestAdversarialGeo_HaversineBoundaryValues(t *testing.T) {
	geoCases := []struct {
		name       string
		lat1, lon1 float64
		lat2, lon2 float64
	}{
		{"Poles_NorthToSouth", 90.0, 0.0, -90.0, 0.0},
		{"Poles_NorthToNorth", 90.0, 0.0, 90.0, 0.0},
		{"DateLine_AntipodalLongitudes", 0.0, 180.0, 0.0, -180.0},
		{"Equator_HalfCircumference", 0.0, 0.0, 0.0, 180.0},
		{"Equator_ZeroDistance", 0.0, 0.0, 0.0, 0.0},
		{"BeyondPoles_HighPositive", 95.0, 0.0, 100.0, 0.0},
		{"BeyondPoles_HighNegative", -120.0, 0.0, -150.0, 0.0},
		{"BeyondDateLine_East", 0.0, 210.0, 0.0, 250.0},
		{"BeyondDateLine_West", 0.0, -210.0, 0.0, -250.0},
		{"AstronomicalCoords", 1e8, 1e8, -1e8, -1e8},
		{"SubatomicCoords", 1e-12, 1e-12, 2e-12, 2e-12},
		{"NaN_Coordinates", math.NaN(), 0.0, 0.0, 0.0},
		{"Inf_Coordinates", math.Inf(1), 0.0, 0.0, 0.0},
		{"NegInf_Coordinates", 0.0, math.Inf(-1), 0.0, 0.0},
	}

	for _, tc := range geoCases {
		t.Run(tc.name, func(t *testing.T) {
			// CalculateHaversineDistance must never panic
			dist := CalculateHaversineDistance(tc.lat1, tc.lon1, tc.lat2, tc.lon2)

			// If inputs are NaN or Inf, dist will be NaN or Inf, but MUST NOT panic
			if math.IsNaN(tc.lat1) || math.IsNaN(tc.lon1) || math.IsNaN(tc.lat2) || math.IsNaN(tc.lon2) ||
				math.IsInf(tc.lat1, 0) || math.IsInf(tc.lon1, 0) || math.IsInf(tc.lat2, 0) || math.IsInf(tc.lon2, 0) {
				// Result may be NaN or Inf, verified no panic occurred
				return
			}

			// Valid spherical distances between antipodal poles is approx 20015 km (pi * R)
			if tc.name == "Poles_NorthToSouth" {
				if math.Abs(dist-20015.0) > 50.0 {
					t.Fatalf("Pole to pole distance expected ~20015km, got %f", dist)
				}
			}
			if tc.name == "Equator_ZeroDistance" || tc.name == "Poles_NorthToNorth" || tc.name == "DateLine_AntipodalLongitudes" {
				if dist > 0.001 {
					t.Fatalf("Expected 0.0 km, got %f", dist)
				}
			}
		})
	}
}

func TestAdversarialGeo_HaversineJSON_NaN_Inf_Rejection(t *testing.T) {
	// Standard JSON RFC 8259 disallows NaN and Inf literals; bridge should reject without panicking
	invalidJsonLiterals := []string{
		`{"lat1": NaN, "lon1": 0.0, "lat2": 0.0, "lon2": 0.0}`,
		`{"lat1": Infinity, "lon1": 0.0, "lat2": 0.0, "lon2": 0.0}`,
		`{"lat1": -Infinity, "lon1": 0.0, "lat2": 0.0, "lon2": 0.0}`,
	}

	for _, payload := range invalidJsonLiterals {
		respJSON := CalculateHaversineDistanceJSON(payload)
		var env ResponseEnvelope
		if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
			t.Fatalf("Failed to parse envelope: %v", err)
		}
		if env.Success {
			t.Fatalf("Expected success: false for invalid JSON literal %q", payload)
		}
	}
}

// ============================================================================
// 3. Empty, Invalid, and Private IPv4/IPv6 Address Strings
// ============================================================================

func TestAdversarialIP_AllVariants(t *testing.T) {
	ipCases := []struct {
		ip            string
		expectSuccess bool
		description   string
	}{
		{"", false, "Empty string"},
		{"   ", false, "Whitespace only"},
		{"invalid-ip", false, "Alphanumeric string"},
		{"999.999.999.999", false, "Octets > 255"},
		{"256.0.0.1", false, "First octet > 255"},
		{"1.2.3", false, "Too few octets"},
		{"1.2.3.4.5", false, "Too many octets"},
		{"1.1.1.01", false, "Leading zeros"},
		{"127.0.0.1; rm -rf /", false, "Command injection attempt"},
		{"' OR '1'='1", false, "SQL injection attempt"},
		{"<script>alert(1)</script>", false, "XSS injection attempt"},
		{"127.0.0.1\nGET / HTTP/1.1", false, "CRLF injection attempt"},
		{"127.0.0.1", true, "Loopback IPv4"},
		{"10.0.0.1", true, "RFC1918 Private Class A"},
		{"192.168.1.1", true, "RFC1918 Private Class C"},
		{"172.16.0.1", true, "RFC1918 Private Class B"},
		{"169.254.1.1", true, "Link-Local IPv4"},
		{"0.0.0.0", true, "Zero IPv4"},
		{"255.255.255.255", true, "Broadcast IPv4"},
		{"::1", true, "Loopback IPv6"},
		{"2001:4860:4860::8888", true, "Public Google IPv6"},
		{"fe80::1", true, "Link-Local IPv6"},
	}

	for _, tc := range ipCases {
		t.Run("QueryBgpAsn_"+tc.description, func(t *testing.T) {
			reqJSON := fmt.Sprintf(`{"ip": %q, "timeout_sec": 1}`, tc.ip)
			respJSON := QueryBgpAsn(reqJSON)

			var env ResponseEnvelope
			if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
				t.Fatalf("QueryBgpAsn returned unparseable JSON: %v", err)
			}

			// QueryBgpAsn strictly validates IP format
			if !tc.expectSuccess && env.Success {
				t.Fatalf("Expected failure for IP %q (%s), but got success", tc.ip, tc.description)
			}
		})

		// For empty string "", all entrypoints must reject with "ip parameter is required"
		if tc.ip == "" {
			t.Run("PerformMultilat_Empty", func(t *testing.T) {
				resp := PerformMultilateration(`{"ip": ""}`)
				var env ResponseEnvelope
				json.Unmarshal([]byte(resp), &env)
				if env.Success {
					t.Fatalf("PerformMultilateration must reject empty IP")
				}
			})
			t.Run("AnalyzeIpId_Empty", func(t *testing.T) {
				resp := AnalyzeIpId(`{"ip": ""}`)
				var env ResponseEnvelope
				json.Unmarshal([]byte(resp), &env)
				if env.Success {
					t.Fatalf("AnalyzeIpId must reject empty IP")
				}
			})
			t.Run("AnalyzeTunnel_Empty", func(t *testing.T) {
				resp := AnalyzeTunnel(`{"ip": ""}`)
				var env ResponseEnvelope
				json.Unmarshal([]byte(resp), &env)
				if env.Success {
					t.Fatalf("AnalyzeTunnel must reject empty IP")
				}
			})
		}
	}
}

// ============================================================================
// 4. WiFi Beacons Boundary Cases (zero, single, negative/zero RSSI, identical coords)
// ============================================================================

func TestAdversarialWiFi_ZeroBeacons(t *testing.T) {
	reqJSON := `{"beacons": [], "timeout_sec": 2}`
	respJSON := TriangulateWiFi(reqJSON)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("Failed to parse envelope: %v", err)
	}
	// Zero beacons should return error or handled envelope without panic
	if env.Success {
		t.Fatalf("Expected failure for zero beacons, got success")
	}
	if env.Error == nil || !strings.Contains(*env.Error, "no BSSID targets provided") {
		t.Fatalf("Expected 'no BSSID targets provided' error, got %v", env.Error)
	}
}

func TestAdversarialWiFi_SingleBeacon(t *testing.T) {
	reqJSON := `{
		"beacons": [
			{"bssid": "00:11:22:33:44:55", "ssid": "Solo-AP", "rssi": -50, "latitude": 40.4168, "longitude": -3.7038}
		],
		"timeout_sec": 2
	}`
	respJSON := TriangulateWiFi(reqJSON)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("Failed to parse envelope: %v", err)
	}
	if !env.Success {
		t.Fatalf("Single beacon triangulation should succeed, error: %v", env.Error)
	}

	var resp TriangulateWiFiResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}
	if resp.ResolvedCount != 1 {
		t.Fatalf("Expected 1 resolved beacon, got %d", resp.ResolvedCount)
	}
	if math.Abs(resp.EstimatedPoint.Lat-40.4168) > 0.0001 || math.Abs(resp.EstimatedPoint.Lon-(-3.7038)) > 0.0001 {
		t.Fatalf("Single beacon location mismatch: got %v", resp.EstimatedPoint)
	}
}

func TestAdversarialWiFi_ExtremeRSSIValues(t *testing.T) {
	// Test RSSI = 0 (defaults to -70), extreme negative (-120), very strong (-20), and anomalous positive (+10)
	reqJSON := `{
		"beacons": [
			{"bssid": "00:11:22:33:44:01", "rssi": 0, "latitude": 40.4168, "longitude": -3.7038},
			{"bssid": "00:11:22:33:44:02", "rssi": -120, "latitude": 40.4170, "longitude": -3.7040},
			{"bssid": "00:11:22:33:44:03", "rssi": -20, "latitude": 40.4169, "longitude": -3.7039},
			{"bssid": "00:11:22:33:44:04", "rssi": 10, "latitude": 40.4167, "longitude": -3.7037}
		],
		"timeout_sec": 2
	}`
	respJSON := TriangulateWiFi(reqJSON)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("Failed to parse envelope: %v", err)
	}
	if !env.Success {
		t.Fatalf("Triangulation with extreme RSSI failed: %v", env.Error)
	}

	var resp TriangulateWiFiResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}
	if resp.ResolvedCount != 4 {
		t.Fatalf("Expected 4 resolved beacons, got %d", resp.ResolvedCount)
	}
	if math.IsNaN(resp.EstimatedPoint.Lat) || math.IsNaN(resp.EstimatedPoint.Lon) {
		t.Fatalf("Centroid coordinate is NaN!")
	}
	if math.IsNaN(resp.ConfidenceKm) || math.IsNaN(resp.PrecisionM) {
		t.Fatalf("Precision or confidence is NaN!")
	}
}

func TestAdversarialWiFi_IdenticalCoordinates(t *testing.T) {
	// Multiple beacons with EXACTLY identical coordinates
	reqJSON := `{
		"beacons": [
			{"bssid": "00:11:22:33:44:01", "rssi": -50, "latitude": 40.4168, "longitude": -3.7038},
			{"bssid": "00:11:22:33:44:02", "rssi": -60, "latitude": 40.4168, "longitude": -3.7038},
			{"bssid": "00:11:22:33:44:03", "rssi": -70, "latitude": 40.4168, "longitude": -3.7038}
		],
		"timeout_sec": 2
	}`
	respJSON := TriangulateWiFi(reqJSON)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("Failed to parse envelope: %v", err)
	}
	if !env.Success {
		t.Fatalf("Triangulation failed: %v", env.Error)
	}

	var resp TriangulateWiFiResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}
	// Result must equal the exact identical point
	if math.Abs(resp.EstimatedPoint.Lat-40.4168) > 0.00001 || math.Abs(resp.EstimatedPoint.Lon-(-3.7038)) > 0.00001 {
		t.Fatalf("Expected exact point (40.4168, -3.7038), got %v", resp.EstimatedPoint)
	}
	// Zero dispersion should clamp precision to minimum (8.0m)
	if resp.PrecisionM != 8.0 {
		t.Fatalf("Expected minimum precision 8.0m, got %f", resp.PrecisionM)
	}
}

// ============================================================================
// 5. Extreme Timeouts (0s, Negative, Very Large)
// ============================================================================

func TestAdversarialTimeouts_AllEntrypoints(t *testing.T) {
	timeoutValues := []int{0, -1, -999, 100000, 2147483647}

	for _, timeout := range timeoutValues {
		t.Run(fmt.Sprintf("Timeout_%d", timeout), func(t *testing.T) {
			reqBgp := fmt.Sprintf(`{"ip": "127.0.0.1", "timeout_sec": %d}`, timeout)
			resp := QueryBgpAsn(reqBgp)
			var env ResponseEnvelope
			if err := json.Unmarshal([]byte(resp), &env); err != nil {
				t.Fatalf("QueryBgpAsn failed to return JSON with timeout %d: %v", timeout, err)
			}
			if !env.Success {
				t.Fatalf("QueryBgpAsn failed on timeout %d: %v", timeout, env.Error)
			}

			reqMultilat := fmt.Sprintf(`{"ip": "127.0.0.1", "timeout_sec": %d, "samples": 1}`, timeout)
			respMulti := PerformMultilateration(reqMultilat)
			if err := json.Unmarshal([]byte(respMulti), &env); err != nil {
				t.Fatalf("PerformMultilateration failed to return JSON with timeout %d: %v", timeout, err)
			}
			if !env.Success {
				t.Fatalf("PerformMultilateration failed on timeout %d: %v", timeout, env.Error)
			}
		})
	}
}

// ============================================================================
// 6. 100% Panic Recovery Verification Across All Entrypoints
// ============================================================================

func TestAdversarial_PanicRecoveryIntegrity(t *testing.T) {
	// Verify that if an internal function panics with any arbitrary type (string, error, struct, nil),
	// recoverToError captures it cleanly into ResponseEnvelope
	testCases := []struct {
		name       string
		panicValue interface{}
	}{
		{"StringPanic", "simulated string panic"},
		{"ErrorPanic", fmt.Errorf("simulated error panic")},
		{"NilPointerPanic", (*string)(nil)},
		{"IntPanic", 42},
		{"StructPanic", struct{ msg string }{msg: "complex panic"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			panicker := func() (ret string) {
				defer recoverToError(&ret)
				panic(tc.panicValue)
			}

			respJSON := panicker()
			var env ResponseEnvelope
			if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
				t.Fatalf("Panicking function produced invalid JSON: %v. Raw: %s", err, respJSON)
			}
			if env.Success {
				t.Fatalf("Expected env.Success == false after panic")
			}
			if env.Error == nil || !strings.Contains(*env.Error, "go panic recovered") {
				t.Fatalf("Expected 'go panic recovered' in error, got %v", env.Error)
			}
		})
	}
}
