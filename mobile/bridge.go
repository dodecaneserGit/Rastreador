package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dodecaneser/rastreador/pkg/ipid"
	"github.com/dodecaneser/rastreador/pkg/l2wifi"
	"github.com/dodecaneser/rastreador/pkg/multilat"
	"github.com/dodecaneser/rastreador/pkg/recon"
	"github.com/dodecaneser/rastreador/pkg/tunnel"
)

// ResponseEnvelope standardizes all JSON return payloads across the bridge
type ResponseEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *string         `json:"error"`
}

func successJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return errorJSON(fmt.Sprintf("serialization error: %v", err))
	}
	env := ResponseEnvelope{
		Success: true,
		Data:    b,
		Error:   nil,
	}
	res, _ := json.Marshal(env)
	return string(res)
}

func errorJSON(msg string) string {
	errStr := msg
	env := ResponseEnvelope{
		Success: false,
		Data:    nil,
		Error:   &errStr,
	}
	res, _ := json.Marshal(env)
	return string(res)
}

func recoverToError(result *string) {
	if r := recover(); r != nil {
		*result = errorJSON(fmt.Sprintf("go panic recovered: %v", r))
	}
}

// ----------------------------------------------------------------------
// 1. QueryBgpAsn
// ----------------------------------------------------------------------

type BgpAsnRequest struct {
	IP            string `json:"ip"`
	TimeoutSec    int    `json:"timeout_sec,omitempty"`
	TimeoutSecAlt int    `json:"timeoutSec,omitempty"`
}

type BgpAsnResponse struct {
	IP          string   `json:"ip"`
	ASN         int      `json:"asn"`
	ASOrg       string   `json:"as_org"`
	Country     string   `json:"country"`
	CountryCode string   `json:"country_code,omitempty"`
	Region      string   `json:"region,omitempty"`
	City        string   `json:"city,omitempty"`
	Zip         string   `json:"zip,omitempty"`
	Facility    string   `json:"facility,omitempty"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	ISP         string   `json:"isp"`
	Hostname    string   `json:"hostname,omitempty"`
	IsCloud     bool     `json:"is_cloud"`
	IsVPN       bool     `json:"is_vpn"`
	IsProxy     bool     `json:"is_proxy"`
	IsTorExit   bool     `json:"is_tor_exit"`
	IsAnycast   bool     `json:"is_anycast"`
	Confidence  string   `json:"confidence"`
	PrecisionKm float64  `json:"precision_km"`
	AirportCode string   `json:"detected_airport_code,omitempty"`
	Indicators  []string `json:"indicators,omitempty"`
}

func QueryBgpAsn(requestJSON string) (ret string) {
	defer recoverToError(&ret)

	var req BgpAsnRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return errorJSON(fmt.Sprintf("invalid request JSON: %v", err))
	}
	if req.IP == "" {
		return errorJSON("ip parameter is required")
	}

	timeout := req.TimeoutSec
	if timeout <= 0 && req.TimeoutSecAlt > 0 {
		timeout = req.TimeoutSecAlt
	}
	if timeout <= 0 {
		timeout = 5
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	info, err := recon.QueryIP(ctx, req.IP)
	if err != nil {
		return errorJSON(err.Error())
	}

	country := info.CountryCode
	if country == "" {
		country = info.Country
	}

	resp := BgpAsnResponse{
		IP:          info.IP,
		ASN:         info.ASN,
		ASOrg:       info.ASOrg,
		Country:     country,
		CountryCode: info.CountryCode,
		Region:      info.Region,
		City:        info.City,
		Zip:         info.Zip,
		Facility:    info.Facility,
		Latitude:    info.Latitude,
		Longitude:   info.Longitude,
		ISP:         info.ISP,
		Hostname:    info.Hostname,
		IsCloud:     info.IsCloud,
		IsVPN:       info.IsVPN,
		IsProxy:     info.IsProxy,
		IsTorExit:   info.IsTorExit,
		IsAnycast:   info.IsAnycast,
		Confidence:  info.Confidence,
		PrecisionKm: info.PrecisionKm,
		AirportCode: info.AirportCode,
		Indicators:  info.Indicators,
	}

	return successJSON(resp)
}

// ----------------------------------------------------------------------
// 2. PerformMultilateration
// ----------------------------------------------------------------------

type MultilatRequest struct {
	IP            string `json:"ip"`
	Ports         []int  `json:"ports,omitempty"`
	Samples       int    `json:"samples,omitempty"`
	RipeKey       string `json:"ripe_key,omitempty"`
	RipeKeyAlt    string `json:"ripeKey,omitempty"`
	RipeProbes    int    `json:"ripe_probes,omitempty"`
	RipeProbesAlt int    `json:"ripeProbes,omitempty"`
	TimeoutSec    int    `json:"timeout_sec,omitempty"`
	TimeoutSecAlt int    `json:"timeoutSec,omitempty"`
}

type MultilatResponse struct {
	EstimatedPoint     multilat.Point     `json:"estimated_point"`
	ConfidenceKm       float64            `json:"confidence_km"`
	ConfidenceRadiusKm float64            `json:"confidence_radius_km,omitempty"`
	UsedLandmarks      []multilat.Landmark `json:"used_landmarks"`
	PolygonBounds      []multilat.Point   `json:"polygon_bounds"`
}

func PerformMultilateration(requestJSON string) (ret string) {
	defer recoverToError(&ret)

	var req MultilatRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return errorJSON(fmt.Sprintf("invalid request JSON: %v", err))
	}
	if req.IP == "" {
		return errorJSON("ip parameter is required")
	}

	timeout := req.TimeoutSec
	if timeout <= 0 && req.TimeoutSecAlt > 0 {
		timeout = req.TimeoutSecAlt
	}
	if timeout <= 0 {
		timeout = 10
	}

	samples := req.Samples
	if samples <= 0 {
		samples = 3
	}

	ports := req.Ports
	if len(ports) == 0 {
		ports = []int{80, 443, 22, 53, 8080}
	}

	ripeKey := req.RipeKey
	if ripeKey == "" && req.RipeKeyAlt != "" {
		ripeKey = req.RipeKeyAlt
	}

	ripeProbes := req.RipeProbes
	if ripeProbes <= 0 && req.RipeProbesAlt > 0 {
		ripeProbes = req.RipeProbesAlt
	}
	if ripeProbes <= 0 {
		ripeProbes = 4
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	reconInfo, _ := recon.QueryIP(ctx, req.IP)
	var regLat, regLon, prec float64
	var metro string
	var isAnycast bool
	if reconInfo != nil {
		regLat = reconInfo.Latitude
		regLon = reconInfo.Longitude
		prec = reconInfo.PrecisionKm
		metro = reconInfo.AirportCode
		isAnycast = reconInfo.IsAnycast
	}

	landmarks, estPt, conf := multilat.PerformMultiVantageProbing(ctx, req.IP, ports, samples, regLat, regLon, prec, metro, isAnycast)

	if ripeKey != "" {
		ripeClient := multilat.NewRIPEAtlasClient(ripeKey)
		ripeLandmarks, err := ripeClient.RunRIPEAtlasProbing(ctx, req.IP, ripeProbes)
		if err == nil && len(ripeLandmarks) > 0 {
			landmarks = append(landmarks, ripeLandmarks...)
			if regLat == 0 && regLon == 0 {
				sol := multilat.SolveCentroidLeastSquares(landmarks)
				estPt = sol.EstimatedPoint
				conf = sol.ConfidenceKm
			}
		}
	}

	resp := MultilatResponse{
		EstimatedPoint:     estPt,
		ConfidenceKm:       conf,
		ConfidenceRadiusKm: conf,
		UsedLandmarks:      landmarks,
		PolygonBounds:      make([]multilat.Point, 0),
	}

	return successJSON(resp)
}

// ----------------------------------------------------------------------
// 3. TriangulateWiFi
// ----------------------------------------------------------------------

type TriangulateWiFiRequest struct {
	Beacons       []WiFiBeaconScanDTO `json:"beacons"`
	WigleKey      string              `json:"wigle_key,omitempty"`
	WigleKeyAlt   string              `json:"wigleKey,omitempty"`
	TimeoutSec    int                 `json:"timeout_sec,omitempty"`
	TimeoutSecAlt int                 `json:"timeoutSec,omitempty"`
}

type WiFiBeaconScanDTO struct {
	BSSID        string  `json:"bssid"`
	SSID         string  `json:"ssid,omitempty"`
	RSSI         int     `json:"rssi,omitempty"`
	Frequency    int     `json:"frequency,omitempty"`
	Channel      int     `json:"channel,omitempty"`
	ChannelWidth int     `json:"channel_width,omitempty"`
	Capabilities string  `json:"capabilities,omitempty"`
	Lat          float64 `json:"latitude,omitempty"`
	Lon          float64 `json:"longitude,omitempty"`
	Timestamp    int64   `json:"timestamp,omitempty"`
}

type TriangulateWiFiResponse struct {
	EstimatedPoint multilat.Point      `json:"estimated_point"`
	ConfidenceKm   float64             `json:"confidence_km"`
	PrecisionM     float64             `json:"precision_meters"`
	ResolvedCount  int                 `json:"resolved_count"`
	TotalBeacons   int                 `json:"total_beacons"`
	StreetAddress  string              `json:"street_address,omitempty"`
	Networks       []WiFiBeaconScanDTO `json:"networks"`
}

func TriangulateWiFi(requestJSON string) (ret string) {
	defer recoverToError(&ret)

	var req TriangulateWiFiRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return errorJSON(fmt.Sprintf("invalid request JSON: %v", err))
	}

	timeout := req.TimeoutSec
	if timeout <= 0 && req.TimeoutSecAlt > 0 {
		timeout = req.TimeoutSecAlt
	}
	if timeout <= 0 {
		timeout = 10
	}

	wigleKey := req.WigleKey
	if wigleKey == "" && req.WigleKeyAlt != "" {
		wigleKey = req.WigleKeyAlt
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	nets := make([]l2wifi.WiFiNetwork, 0, len(req.Beacons))
	for _, b := range req.Beacons {
		nets = append(nets, l2wifi.WiFiNetwork{
			BSSID:    b.BSSID,
			SSID:     b.SSID,
			RSSI:     b.RSSI,
			Channel:  b.Channel,
			Lat:      b.Lat,
			Lon:      b.Lon,
			Resolved: (b.Lat != 0 || b.Lon != 0),
		})
	}

	wigleClient := l2wifi.NewWiGLEClient(wigleKey)
	res, err := wigleClient.TriangulateBSSIDs(ctx, nets)
	if err != nil && (res == nil || res.ResolvedCount == 0) {
		return errorJSON(err.Error())
	}

	outNets := make([]WiFiBeaconScanDTO, 0, len(res.Networks))
	for i, n := range res.Networks {
		var orig WiFiBeaconScanDTO
		if i < len(req.Beacons) {
			orig = req.Beacons[i]
		}
		orig.BSSID = n.BSSID
		orig.SSID = n.SSID
		orig.RSSI = n.RSSI
		orig.Lat = n.Lat
		orig.Lon = n.Lon
		outNets = append(outNets, orig)
	}

	resp := TriangulateWiFiResponse{
		EstimatedPoint: res.EstimatedPoint,
		ConfidenceKm:   res.ConfidenceKm,
		PrecisionM:     res.PrecisionM,
		ResolvedCount:  res.ResolvedCount,
		TotalBeacons:   res.TotalBeacons,
		StreetAddress:  res.StreetAddress,
		Networks:       outNets,
	}

	return successJSON(resp)
}

// ----------------------------------------------------------------------
// 4. AnalyzeIpId
// ----------------------------------------------------------------------

type AnalyzeIpIdRequest struct {
	IP            string `json:"ip"`
	Ports         []int  `json:"ports,omitempty"`
	Samples       int    `json:"samples,omitempty"`
	TimeoutSec    int    `json:"timeout_sec,omitempty"`
	TimeoutSecAlt int    `json:"timeoutSec,omitempty"`
}

type AnalyzeIpIdResponse struct {
	TargetIP               string          `json:"target_ip"`
	Timestamp              string          `json:"timestamp"`
	GenerationType         string          `json:"generation_type"`
	VelocityPacketsPerSec  float64         `json:"velocity_packets_per_sec"`
	ClockFrequencyHz       float64         `json:"clock_frequency_hz,omitempty"`
	LinearityScore         float64         `json:"linearity_score"`
	CorrelationFingerprint string          `json:"correlation_fingerprint"`
	TotalSamples           int             `json:"total_samples"`
	Samples                []IPIDSampleDTO `json:"samples"`
	Findings               []string        `json:"findings"`
}

type IPIDSampleDTO struct {
	TimestampMs int64   `json:"timestamp_ms"`
	DeltaMs     float64 `json:"delta_ms"`
	IPID        int     `json:"ip_id"`
	DeltaID     int     `json:"delta_id"`
	TCPTSVal    int64   `json:"tcp_tsval,omitempty"`
	DeltaTS     int64   `json:"delta_tsval,omitempty"`
}

func AnalyzeIpId(requestJSON string) (ret string) {
	defer recoverToError(&ret)

	var req AnalyzeIpIdRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return errorJSON(fmt.Sprintf("invalid request JSON: %v", err))
	}
	if req.IP == "" {
		return errorJSON("ip parameter is required")
	}

	timeout := req.TimeoutSec
	if timeout <= 0 && req.TimeoutSecAlt > 0 {
		timeout = req.TimeoutSecAlt
	}
	if timeout <= 0 {
		timeout = 5
	}

	samples := req.Samples
	if samples <= 0 {
		samples = 6
	}

	ports := req.Ports
	if len(ports) == 0 {
		ports = []int{80, 443, 22, 53, 8080}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	res, err := ipid.AnalyzeIPID(ctx, req.IP, ports, samples)
	if err != nil {
		return errorJSON(err.Error())
	}

	sampleDTOs := make([]IPIDSampleDTO, 0, len(res.Samples))
	for _, s := range res.Samples {
		sampleDTOs = append(sampleDTOs, IPIDSampleDTO{
			TimestampMs: s.Timestamp.UnixMilli(),
			DeltaMs:     s.DeltaMs,
			IPID:        int(s.IPID),
			DeltaID:     s.DeltaID,
			TCPTSVal:    int64(s.TCPTSVal),
			DeltaTS:     s.DeltaTS,
		})
	}

	resp := AnalyzeIpIdResponse{
		TargetIP:               res.TargetIP,
		Timestamp:              res.Timestamp,
		GenerationType:         res.GenerationType,
		VelocityPacketsPerSec:  res.VelocityPacketsPerSec,
		ClockFrequencyHz:       res.ClockFrequencyHz,
		LinearityScore:         res.LinearityScore,
		CorrelationFingerprint: res.CorrelationFingerprint,
		TotalSamples:           res.TotalSamples,
		Samples:                sampleDTOs,
		Findings:               res.Findings,
	}

	return successJSON(resp)
}

// ----------------------------------------------------------------------
// 5. AnalyzeTunnel
// ----------------------------------------------------------------------

type AnalyzeTunnelRequest struct {
	IP            string `json:"ip"`
	L7URL         string `json:"l7_url,omitempty"`
	L7URLAlt      string `json:"l7Url,omitempty"`
	TimeoutSec    int    `json:"timeout_sec,omitempty"`
	TimeoutSecAlt int    `json:"timeoutSec,omitempty"`
}

type AnalyzeTunnelResponse struct {
	TargetIP            string   `json:"target_ip"`
	IsTunnelDetected    bool     `json:"is_tunnel_detected"`
	TunnelType          string   `json:"tunnel_type"`
	MssClampingDetected bool     `json:"mss_clamping_detected"`
	LatencyInflationMs  float64  `json:"latency_inflation_ms"`
	Confidence          string   `json:"confidence"`
	L4TCPRTTMs          float64  `json:"l4_tcp_rtt_ms,omitempty"`
	L7AppRTTMs          float64  `json:"l7_app_rtt_ms,omitempty"`
	DeltaRTTMs          float64  `json:"delta_rtt_ms,omitempty"`
	EstimatedTunnelKm   float64  `json:"estimated_tunnel_km,omitempty"`
	Findings            []string `json:"findings,omitempty"`
}

func AnalyzeTunnel(requestJSON string) (ret string) {
	defer recoverToError(&ret)

	var req AnalyzeTunnelRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return errorJSON(fmt.Sprintf("invalid request JSON: %v", err))
	}
	if req.IP == "" {
		return errorJSON("ip parameter is required")
	}

	timeout := req.TimeoutSec
	if timeout <= 0 && req.TimeoutSecAlt > 0 {
		timeout = req.TimeoutSecAlt
	}
	if timeout <= 0 {
		timeout = 5
	}

	l7URL := req.L7URL
	if l7URL == "" && req.L7URLAlt != "" {
		l7URL = req.L7URLAlt
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	res, err := tunnel.AnalyzeTunnel(ctx, req.IP, l7URL)
	if err != nil {
		return errorJSON(err.Error())
	}

	tunnelType := res.TunnelTypeGuess
	if tunnelType == "" {
		if res.HasTunnelOverhead {
			tunnelType = "WIREGUARD"
		} else {
			tunnelType = "NONE"
		}
	}

	mssClamped := res.MTUEstimate > 0 && res.MTUEstimate < 1420
	conf := "LOW"
	if res.HasTunnelOverhead && res.DeltaRTTMs > 20.0 {
		conf = "HIGH"
	}

	resp := AnalyzeTunnelResponse{
		TargetIP:            res.TargetIP,
		IsTunnelDetected:    res.HasTunnelOverhead,
		TunnelType:          tunnelType,
		MssClampingDetected: mssClamped,
		LatencyInflationMs:  res.DeltaRTTMs,
		Confidence:          conf,
		L4TCPRTTMs:          res.L4TCPRTTMs,
		L7AppRTTMs:          res.L7AppRTTMs,
		DeltaRTTMs:          res.DeltaRTTMs,
		EstimatedTunnelKm:   res.EstimatedTunnelKm,
		Findings:            res.Findings,
	}

	return successJSON(resp)
}

// ----------------------------------------------------------------------
// 6. CalculateHaversineDistance & JSON wrapper
// ----------------------------------------------------------------------

func CalculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	return multilat.DistanceHaversine(
		multilat.Point{Lat: lat1, Lon: lon1},
		multilat.Point{Lat: lat2, Lon: lon2},
	)
}

type HaversineRequest struct {
	Lat1 float64 `json:"lat1"`
	Lon1 float64 `json:"lon1"`
	Lat2 float64 `json:"lat2"`
	Lon2 float64 `json:"lon2"`
}

type HaversineResponse struct {
	DistanceKm float64 `json:"distance_km"`
}

func CalculateHaversineDistanceJSON(requestJSON string) (ret string) {
	defer recoverToError(&ret)

	var req HaversineRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return errorJSON(fmt.Sprintf("invalid request JSON: %v", err))
	}
	dist := CalculateHaversineDistance(req.Lat1, req.Lon1, req.Lat2, req.Lon2)
	return successJSON(HaversineResponse{DistanceKm: dist})
}

// ----------------------------------------------------------------------
// Multi-Action Multiplexer (InvokeJson)
// ----------------------------------------------------------------------

func InvokeJson(action string, payloadJSON string) string {
	switch action {
	case "queryBgpAsn":
		return QueryBgpAsn(payloadJSON)
	case "performMultilateration":
		return PerformMultilateration(payloadJSON)
	case "triangulateWiFi":
		return TriangulateWiFi(payloadJSON)
	case "analyzeIpId":
		return AnalyzeIpId(payloadJSON)
	case "analyzeTunnel":
		return AnalyzeTunnel(payloadJSON)
	case "calculateHaversineDistance":
		return CalculateHaversineDistanceJSON(payloadJSON)
	default:
		return errorJSON(fmt.Sprintf("unknown action: %s", action))
	}
}
