package recon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IPInfo holds complete classification and infrastructure details
type IPInfo struct {
	IP          string   `json:"ip"`
	Hostname    string   `json:"hostname,omitempty"`
	ASN         int      `json:"asn"`
	ASOrg       string   `json:"as_org"`
	Country     string   `json:"country"`
	CountryCode string   `json:"country_code"`
	Region      string   `json:"region,omitempty"`
	City        string   `json:"city,omitempty"`
	Zip         string   `json:"zip,omitempty"`
	Facility    string   `json:"facility,omitempty"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	ISP         string   `json:"isp"`
	IsCloud     bool     `json:"is_cloud"`
	IsVPN       bool     `json:"is_vpn"`
	IsProxy     bool     `json:"is_proxy"`
	IsTorExit   bool     `json:"is_tor_exit"`
	IsAnycast   bool     `json:"is_anycast"`
	Confidence  string   `json:"confidence"`
	PrecisionKm float64  `json:"precision_km"`
	AirportCode string   `json:"detected_airport_code,omitempty"`
	Indicators  []string `json:"indicators"`
}

// Known Datacenter/Hosting/VPN ASNs and names
var knownCloudKeywords = []string{
	"amazon", "aws", "google", "digitalocean", "hetzner", "ovh", "linode", "vultr",
	"leaseweb", "oracle", "microsoft", "azure", "cloudflare", "fastly", "m247",
	"datacamp", "choopa", "akamai", "contabo", "scaleaway", "packet", "equinix",
	"sakura", "yourserver", "server", "hosting", "cloud",
}

var knownVPNKeywords = []string{
	"mullvad", "nordvpn", "expressvpn", "surfshark", "protonvpn", "cyberghost",
	"ipvanish", "privateinternetaccess", "windscribe", "torguard", "hidemyass",
	"vpn", "proxy", "tor-exit", "exit-node", "relay", "datacamp", "m247",
}

var knownAnycastASNs = map[int]string{
	13335: "Cloudflare Anycast DNS / CDN",
	15169: "Google Anycast Infrastructure",
	19281: "Quad9 Anycast DNS",
	20940: "Akamai Anycast CDN",
	54113: "Fastly Anycast CDN",
	26347: "DreamHost Anycast",
	16509: "Amazon CloudFront Anycast",
}

// Airport / Metro code regex for PTR hostnames (e.g., fra02s21-in-f14.1e100.net, mad01-vpn.net)
var iataRegex = regexp.MustCompile(`(?i)\b(mad|bcn|fra|ams|lhr|cdg|iad|sfo|lax|ord|dfw|jfk|ewr|nrt|hnd|sin|syd|gru|zrh|vie|arn|hel|cph|dub|mil|mxp|lin|fco)\b`)

// QueryIP performs multi-source reconnaissance on the given IP address
func QueryIP(ctx context.Context, ipStr string) (*IPInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	info := &IPInfo{
		IP:          ipStr,
		PrecisionKm: 1.0, // default target precision
		Indicators:  make([]string, 0),
	}

	// 1. Reverse PTR Lookup
	ptrs, err := net.LookupAddr(ipStr)
	if err == nil && len(ptrs) > 0 {
		info.Hostname = strings.TrimSuffix(ptrs[0], ".")
		info.Indicators = append(info.Indicators, fmt.Sprintf("PTR: %s", info.Hostname))

		// Check for airport/metro code in PTR
		if match := iataRegex.FindString(info.Hostname); match != "" {
			info.AirportCode = strings.ToUpper(match)
			info.Indicators = append(info.Indicators, fmt.Sprintf("IATA Metro Code in PTR: %s", info.AirportCode))
		}
	}

	// 2. Primary Geolocation & ISP Enrichment (ALWAYS called)
	enrichViaIPAPI(ctx, ipStr, info)

	// Fallback to secondary API if coordinates are still 0
	if info.Latitude == 0 && info.Longitude == 0 {
		enrichViaIPWhois(ctx, ipStr, info)
	}

	// 3. Query BGP ASN via Team Cymru DNS
	queryCymruASN(ip, info)

	// 4. Ground-Truth Facility & Campus Match (Sub-1km Precision)
	if fac := FindMatchingFacility(info.ASN, info.ASOrg, info.ISP, info.Hostname, info.City, info.Zip, info.Country); fac != nil {
		info.Facility = fac.Name
		info.Latitude = fac.Lat
		info.Longitude = fac.Lon
		info.PrecisionKm = fac.PrecisionKm
		info.Indicators = append(info.Indicators, fmt.Sprintf("Ground-Truth Facility: %s (±%.2f km)", fac.Name, fac.PrecisionKm))
	} else if info.Zip != "" {
		if lat, lon, prec, ok := PostalCentroid(info.CountryCode, info.Zip); ok {
			info.Latitude = lat
			info.Longitude = lon
			info.PrecisionKm = prec
			info.Indicators = append(info.Indicators, fmt.Sprintf("Postal Centroid [%s]: ±%.2f km", info.Zip, prec))
		}
	}

	// 5. Heuristic classification & Anycast detection
	classifyIP(info)

	return info, nil
}

type ipAPISummary struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Hosting     bool    `json:"hosting"`
	Proxy       bool    `json:"proxy"`
}

func enrichViaIPAPI(ctx context.Context, ipStr string, info *IPInfo) {
	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,regionName,city,zip,lat,lon,isp,org,as,hosting,proxy", ipStr), nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var data ipAPISummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Status != "success" {
		return
	}

	info.Country = data.Country
	info.CountryCode = data.CountryCode
	info.Region = data.RegionName
	info.City = data.City
	info.Zip = data.Zip
	info.Latitude = data.Lat
	info.Longitude = data.Lon

	if data.ISP != "" {
		info.ISP = data.ISP
	}
	if data.Org != "" {
		info.ASOrg = data.Org
	}
	if data.Hosting {
		info.IsCloud = true
	}
	if data.Proxy {
		info.IsProxy = true
	}

	// Parse ASN number if present in "AS15169 Google LLC"
	if strings.HasPrefix(data.AS, "AS") {
		parts := strings.Fields(data.AS)
		if len(parts) > 0 {
			asnStr := strings.TrimPrefix(parts[0], "AS")
			if num, err := strconv.Atoi(asnStr); err == nil && info.ASN == 0 {
				info.ASN = num
			}
		}
	}
}

type ipWhoisSummary struct {
	Success bool    `json:"success"`
	Country string  `json:"country"`
	City    string  `json:"city"`
	Region  string  `json:"region"`
	Lat     float64 `json:"latitude"`
	Lon     float64 `json:"longitude"`
	ISP     string  `json:"isp"`
	Org     string  `json:"org"`
	ASN     int     `json:"asn"`
}

func enrichViaIPWhois(ctx context.Context, ipStr string, info *IPInfo) {
	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://ipwho.is/%s", ipStr), nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var data ipWhoisSummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || !data.Success {
		return
	}

	if info.Country == "" {
		info.Country = data.Country
	}
	if info.City == "" {
		info.City = data.City
	}
	if info.Region == "" {
		info.Region = data.Region
	}
	info.Latitude = data.Lat
	info.Longitude = data.Lon
	if info.ISP == "" {
		info.ISP = data.ISP
	}
	if info.ASOrg == "" {
		info.ASOrg = data.Org
	}
	if info.ASN == 0 {
		info.ASN = data.ASN
	}
}

func queryCymruASN(ip net.IP, info *IPInfo) {
	var reverseQuery string
	if ip4 := ip.To4(); ip4 != nil {
		reverseQuery = fmt.Sprintf("%d.%d.%d.%d.origin.asn.cymru.com", ip4[3], ip4[2], ip4[1], ip4[0])
	} else {
		return
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 2 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	txts, err := r.LookupTXT(context.Background(), reverseQuery)
	if err != nil || len(txts) == 0 {
		return
	}

	// Format: "13335 | 1.1.1.0/24 | US | arin | 2014-04-03"
	parts := strings.Split(txts[0], "|")
	if len(parts) >= 3 {
		asnStr := strings.TrimSpace(parts[0])
		if asn, err := strconv.Atoi(asnStr); err == nil && info.ASN == 0 {
			info.ASN = asn
		}
		if info.Country == "" {
			info.Country = strings.TrimSpace(parts[2])
		}
	}

	// Also query AS Org description if empty
	if info.ASN != 0 && info.ASOrg == "" {
		asQuery := fmt.Sprintf("AS%d.asn.cymru.com", info.ASN)
		asTxts, err := r.LookupTXT(context.Background(), asQuery)
		if err == nil && len(asTxts) > 0 {
			asParts := strings.Split(asTxts[0], "|")
			if len(asParts) >= 5 {
				info.ASOrg = strings.TrimSpace(asParts[4])
			}
		}
	}
}

func classifyIP(info *IPInfo) {
	// Check Anycast database
	if desc, ok := knownAnycastASNs[info.ASN]; ok {
		info.IsAnycast = true
		info.Indicators = append(info.Indicators, fmt.Sprintf("BGP Anycast detected: %s", desc))
	}

	combined := strings.ToLower(fmt.Sprintf("%s %s %s %s", info.Hostname, info.ASOrg, info.ISP, info.Country))

	// Check Cloud keywords
	for _, kw := range knownCloudKeywords {
		if strings.Contains(combined, kw) {
			info.IsCloud = true
			info.Indicators = append(info.Indicators, fmt.Sprintf("Cloud/Datacenter provider match: '%s'", kw))
			break
		}
	}

	// Check VPN keywords
	for _, kw := range knownVPNKeywords {
		if strings.Contains(combined, kw) {
			info.IsVPN = true
			info.Indicators = append(info.Indicators, fmt.Sprintf("Commercial VPN/Proxy signature match: '%s'", kw))
			break
		}
	}

	// Set Confidence
	if info.IsAnycast {
		info.Confidence = "BGP_ANYCAST_EDGE"
	} else if info.IsVPN || info.IsProxy {
		info.Confidence = "MASKED_VPN_PROXY"
	} else if info.IsCloud {
		info.Confidence = "DATACENTER_EXIT"
	} else {
		info.Confidence = "RESIDENTIAL_OR_DIRECT"
	}
}
