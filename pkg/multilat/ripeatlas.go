package multilat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// RIPEAtlasClient handles interactions with the official RIPE Atlas API v2
type RIPEAtlasClient struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewRIPEAtlasClient creates a client configured with API key if provided, env var, or zshrc
func NewRIPEAtlasClient(apiKey string) *RIPEAtlasClient {
	if apiKey == "" {
		apiKey = os.Getenv("RIPE_ATLAS_KEY")
	}
	if apiKey == "" {
		apiKey = findKeyInConfigFiles()
	}
	return &RIPEAtlasClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func findKeyInConfigFiles() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	files := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".config", "rastreador", "config.env"),
	}

	keyRegex := regexp.MustCompile(`(?i)(?:export\s+)?RIPE_ATLAS_KEY=["']?([a-f0-9\-]{30,})["']?`)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err == nil {
			matches := keyRegex.FindStringSubmatch(string(data))
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1])
			}
		}
	}
	return ""
}

type ripeMeasurementRequest struct {
	Definitions []ripeDefinition `json:"definitions"`
	Probes      []ripeProbeSpec  `json:"probes"`
}

type ripeDefinition struct {
	Type        string `json:"type"`
	AF          int    `json:"af"`
	Target      string `json:"target"`
	Description string `json:"description"`
	Packets     int    `json:"packets"`
	IsOneOff    bool   `json:"is_oneoff"`
}

type ripeProbeSpec struct {
	Requested int    `json:"requested"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

type ripeCreateResponse struct {
	Measurements []int `json:"measurements"`
	Error        struct {
		Detail string `json:"detail"`
		Errors []struct {
			Detail string `json:"detail"`
		} `json:"errors"`
	} `json:"error"`
}

type ripeResultItem struct {
	PrbID   int     `json:"prb_id"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	SrcAddr string  `json:"src_addr"`
}

type ripeProbeDetail struct {
	ID          int    `json:"id"`
	CountryCode string `json:"country_code"`
	City        string `json:"city,omitempty"`
	Geometry    struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"` // [lon, lat]
	} `json:"geometry"`
	IsAnchor bool `json:"is_anchor"`
}

// RunRIPEAtlasProbing orchestrates global probes using RIPE Atlas
func (c *RIPEAtlasClient) RunRIPEAtlasProbing(ctx context.Context, targetIP string, probeCount int) ([]Landmark, error) {
	if c.APIKey == "" {
		return c.FetchPublicMeasurements(ctx, targetIP)
	}

	if probeCount <= 0 {
		probeCount = 4
	}

	// 1. Create One-Off Ping Measurement
	reqBody := ripeMeasurementRequest{
		Definitions: []ripeDefinition{
			{
				Type:        "ping",
				AF:          4,
				Target:      targetIP,
				Description: "Rastreador Multilateration Probe",
				Packets:     3,
				IsOneOff:    true,
			},
		},
		Probes: []ripeProbeSpec{
			{
				Requested: probeCount,
				Type:      "area",
				Value:     "WW",
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://atlas.ripe.net/api/v2/measurements/", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Key "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error creating RIPE Atlas measurement: %w", err)
	}
	defer resp.Body.Close()

	var createResp ripeCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil || len(createResp.Measurements) == 0 {
		detail := createResp.Error.Detail
		if len(createResp.Error.Errors) > 0 && createResp.Error.Errors[0].Detail != "" {
			detail = createResp.Error.Errors[0].Detail
		}
		if detail == "" {
			detail = "error scheduling measurement in RIPE Atlas"
		}
		return nil, fmt.Errorf("%s", detail)
	}

	measurementID := createResp.Measurements[0]

	// 2. Poll for results (give RIPE probes up to 18 seconds)
	var results []ripeResultItem
	for i := 0; i < 7; i++ {
		time.Sleep(3 * time.Second)
		results, err = c.getMeasurementResults(ctx, measurementID)
		if err == nil && len(results) >= 1 {
			// Check if at least one probe got a valid response (min > 0)
			hasValid := false
			for _, r := range results {
				if r.Min > 0 || r.Avg > 0 {
					hasValid = true
					break
				}
			}
			if hasValid {
				break
			}
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("target host did not respond to ICMP probes or timed out")
	}

	// 3. Resolve Probe Coordinates and construct Landmarks
	return c.convertResultsToLandmarks(ctx, results)
}

func (c *RIPEAtlasClient) getMeasurementResults(ctx context.Context, measurementID int) ([]ripeResultItem, error) {
	url := fmt.Sprintf("https://atlas.ripe.net/api/v2/measurements/%d/results/", measurementID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var items []ripeResultItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}

// FetchPublicMeasurements retrieves existing public measurements on the target IP if available
func (c *RIPEAtlasClient) FetchPublicMeasurements(ctx context.Context, targetIP string) ([]Landmark, error) {
	url := fmt.Sprintf("https://atlas.ripe.net/api/v2/measurements/?target=%s&type=ping&status=4&page_size=3", targetIP)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no public RIPE Atlas measurements available")
	}
	defer resp.Body.Close()

	var searchResp struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil || len(searchResp.Results) == 0 {
		return nil, fmt.Errorf("no public measurement records found for %s", targetIP)
	}

	latestID := searchResp.Results[0].ID
	results, err := c.getMeasurementResults(ctx, latestID)
	if err != nil || len(results) == 0 {
		return nil, err
	}

	return c.convertResultsToLandmarks(ctx, results)
}

func (c *RIPEAtlasClient) convertResultsToLandmarks(ctx context.Context, results []ripeResultItem) ([]Landmark, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	landmarks := make([]Landmark, 0)

	for _, item := range results {
		if item.Min <= 0 && item.Avg <= 0 {
			continue
		}
		rtt := item.Min
		if rtt <= 0 {
			rtt = item.Avg
		}

		wg.Add(1)
		go func(prbID int, minRTT float64) {
			defer wg.Done()
			probeDetail, err := c.getProbeDetail(ctx, prbID)
			if err != nil || len(probeDetail.Geometry.Coordinates) < 2 {
				return
			}

			lon := probeDetail.Geometry.Coordinates[0]
			lat := probeDetail.Geometry.Coordinates[1]

			lm := Landmark{
				ID:        fmt.Sprintf("RIPE-PRB-%d", prbID),
				Name:      fmt.Sprintf("RIPE Sonda #%d (%s)", prbID, probeDetail.CountryCode),
				City:      probeDetail.CountryCode,
				Country:   probeDetail.CountryCode,
				Location:  Point{Lat: lat, Lon: lon},
				MinRTT:    minRTT,
				MaxRadius: ConstraintRadiusFromRTT(minRTT),
				Samples:   3,
				Type:      "ripe_atlas_probe",
			}

			mu.Lock()
			landmarks = append(landmarks, lm)
			mu.Unlock()
		}(item.PrbID, rtt)
	}

	wg.Wait()
	return landmarks, nil
}

func (c *RIPEAtlasClient) getProbeDetail(ctx context.Context, probeID int) (*ripeProbeDetail, error) {
	url := fmt.Sprintf("https://atlas.ripe.net/api/v2/probes/%d/", probeID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not fetch probe detail")
	}
	defer resp.Body.Close()

	var detail ripeProbeDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}

	return &detail, nil
}
