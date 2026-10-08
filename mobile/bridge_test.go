package mobile

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestCalculateHaversineDistance(t *testing.T) {
	// Madrid (40.4168, -3.7038) to Barcelona (41.3851, 2.1734) is approx 505 km
	dist := CalculateHaversineDistance(40.4168, -3.7038, 41.3851, 2.1734)
	if math.Abs(dist-505.0) > 10.0 {
		t.Fatalf("expected distance ~505km, got %f", dist)
	}

	// Zero distance
	zero := CalculateHaversineDistance(40.4168, -3.7038, 40.4168, -3.7038)
	if zero != 0.0 {
		t.Fatalf("expected 0.0, got %f", zero)
	}
}

func TestCalculateHaversineDistanceJSON(t *testing.T) {
	req := `{"lat1": 40.4168, "lon1": -3.7038, "lat2": 41.3851, "lon2": 2.1734}`
	respJSON := CalculateHaversineDistanceJSON(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true, got false")
	}

	var data HaversineResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if math.Abs(data.DistanceKm-505.0) > 10.0 {
		t.Fatalf("expected ~505km, got %f", data.DistanceKm)
	}
}

func TestQueryBgpAsn_ValidIP(t *testing.T) {
	req := `{"ip": "147.96.1.1", "timeout_sec": 3}`
	respJSON := QueryBgpAsn(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true, error: %v", env.Error)
	}

	var resp BgpAsnResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal BgpAsnResponse: %v", err)
	}
	if resp.IP != "147.96.1.1" {
		t.Fatalf("expected IP 147.96.1.1, got %s", resp.IP)
	}
	if resp.ASN != 766 {
		t.Logf("ASN is %d (expected 766 if offline or facility matched)", resp.ASN)
	}
}

func TestQueryBgpAsn_InvalidIP(t *testing.T) {
	req := `{"ip": "invalid-ip"}`
	respJSON := QueryBgpAsn(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if env.Success {
		t.Fatalf("expected success false for invalid IP")
	}
	if env.Error == nil || !strings.Contains(*env.Error, "invalid IP") {
		t.Fatalf("expected invalid IP error, got %v", env.Error)
	}
}

func TestQueryBgpAsn_EmptyIP(t *testing.T) {
	req := `{"ip": ""}`
	respJSON := QueryBgpAsn(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if env.Success {
		t.Fatalf("expected success false for empty IP")
	}
}

func TestPerformMultilateration_Valid(t *testing.T) {
	req := `{"ip": "127.0.0.1", "timeout_sec": 3, "samples": 2}`
	respJSON := PerformMultilateration(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true, got %v", env.Error)
	}

	var resp MultilatResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal MultilatResponse: %v", err)
	}
	if resp.ConfidenceKm <= 0 {
		t.Fatalf("expected positive confidence radius, got %f", resp.ConfidenceKm)
	}
}

func TestTriangulateWiFi_SyntheticBeacons(t *testing.T) {
	req := `{
		"beacons": [
			{"bssid": "00:11:22:33:44:55", "ssid": "Test-AP1", "rssi": -55, "latitude": 40.4168, "longitude": -3.7038},
			{"bssid": "00:11:22:33:44:66", "ssid": "Test-AP2", "rssi": -65, "latitude": 40.4170, "longitude": -3.7040}
		],
		"timeout_sec": 5
	}`
	respJSON := TriangulateWiFi(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true, got %v", env.Error)
	}

	var resp TriangulateWiFiResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal TriangulateWiFiResponse: %v", err)
	}
	if resp.ResolvedCount != 2 {
		t.Fatalf("expected 2 resolved beacons, got %d", resp.ResolvedCount)
	}
	if resp.EstimatedPoint.Lat == 0.0 || resp.EstimatedPoint.Lon == 0.0 {
		t.Fatalf("expected non-zero estimated point, got %v", resp.EstimatedPoint)
	}
}

func TestAnalyzeIpId_Valid(t *testing.T) {
	req := `{"ip": "127.0.0.1", "samples": 4, "timeout_sec": 3}`
	respJSON := AnalyzeIpId(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true, got error: %v", env.Error)
	}

	var resp AnalyzeIpIdResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal AnalyzeIpIdResponse: %v", err)
	}
	if resp.TargetIP != "127.0.0.1" {
		t.Fatalf("expected target 127.0.0.1, got %s", resp.TargetIP)
	}
	if resp.GenerationType == "" {
		t.Fatalf("expected non-empty generation type")
	}
}

func TestAnalyzeTunnel_Valid(t *testing.T) {
	req := `{"ip": "127.0.0.1", "timeout_sec": 3}`
	respJSON := AnalyzeTunnel(req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true, got %v", env.Error)
	}

	var resp AnalyzeTunnelResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("failed to unmarshal AnalyzeTunnelResponse: %v", err)
	}
	if resp.TargetIP != "127.0.0.1" {
		t.Fatalf("expected target 127.0.0.1, got %s", resp.TargetIP)
	}
	if resp.TunnelType == "" {
		t.Fatalf("expected non-empty tunnel type")
	}
}

func TestPanicRecovery(t *testing.T) {
	panickyFunc := func() (ret string) {
		defer recoverToError(&ret)
		panic("intentional simulated panic")
	}

	respJSON := panickyFunc()
	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if env.Success {
		t.Fatalf("expected success false after panic")
	}
	if env.Error == nil || !strings.Contains(*env.Error, "simulated panic") {
		t.Fatalf("expected panic message in error, got %v", env.Error)
	}
}

func TestInvokeJson(t *testing.T) {
	req := `{"lat1": 0, "lon1": 0, "lat2": 0, "lon2": 1}`
	respJSON := InvokeJson("calculateHaversineDistance", req)

	var env ResponseEnvelope
	if err := json.Unmarshal([]byte(respJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success true")
	}

	unknownJSON := InvokeJson("unknownAction", "{}")
	if !strings.Contains(unknownJSON, "unknown action") {
		t.Fatalf("expected unknown action error, got %s", unknownJSON)
	}
}
