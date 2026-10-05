package recon

import (
	"net"
)

// CentralRecord represents a physical telecom central office (Central Telefónica / BAP / BRAS)
type CentralRecord struct {
	ID          string  `json:"id"`
	ISP         string  `json:"isp"` // "Movistar / Telefónica", "Orange", "Vodafone", "Digi"
	Name        string  `json:"name"`
	District    string  `json:"district"`
	City        string  `json:"city"`
	Province    string  `json:"province"`
	CountryCode string  `json:"country_code"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	PrecisionKm float64 `json:"precision_km"` // Typically 1.0 - 1.5 km (neighborhood coverage radius)
	CIDRs       []string`json:"cidrs"`
}

// Master database of Spanish telecom central offices and sub-allocation blocks (Telefónica RIMA, Orange, Vodafone)
var knownCentrales = []CentralRecord{
	// ==================== MADRID CAPITAL Y ÁREA METROPOLITANA ====================
	{
		ID: "ES-MAD-TETUAN", ISP: "Telefónica de España (RIMA)",
		Name: "Central Tetuán / Cuatro Caminos (Bravo Murillo)", District: "Tetuán / Cuatro Caminos",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.4535, Lon: -3.7042, PrecisionKm: 1.2,
		CIDRs: []string{"2.139.24.0/22", "2.139.25.0/24", "80.58.12.0/22", "81.36.48.0/21", "83.34.128.0/20"},
	},
	{
		ID: "ES-MAD-CHAMBERI", ISP: "Telefónica de España (RIMA)",
		Name: "Central Chamberí (Santa Engracia)", District: "Chamberí",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.4370, Lon: -3.6990, PrecisionKm: 1.0,
		CIDRs: []string{"2.139.20.0/22", "80.58.16.0/22", "83.34.64.0/20", "88.2.128.0/20"},
	},
	{
		ID: "ES-MAD-CENTRO", ISP: "Telefónica de España (RIMA)",
		Name: "Central San Jerónimo (Centro Histórico / Gran Vía)", District: "Centro / Sol",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.4165, Lon: -3.7010, PrecisionKm: 1.1,
		CIDRs: []string{"2.139.16.0/22", "80.58.0.0/22", "81.36.32.0/21", "83.34.0.0/20"},
	},
	{
		ID: "ES-MAD-SALAMANCA", ISP: "Telefónica de España (RIMA)",
		Name: "Central Goya / Salamanca (Calle Goya / Serrano)", District: "Barrio de Salamanca",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.4250, Lon: -3.6820, PrecisionKm: 1.2,
		CIDRs: []string{"2.139.28.0/22", "80.58.20.0/22", "83.34.96.0/20", "88.2.160.0/20"},
	},
	{
		ID: "ES-MAD-CHAMARTIN", ISP: "Telefónica de España (RIMA)",
		Name: "Central Chamartín (Mateo Inurria / Plaza Castilla)", District: "Chamartín / Castilla",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.4680, Lon: -3.6870, PrecisionKm: 1.3,
		CIDRs: []string{"2.139.32.0/22", "80.58.24.0/22", "83.35.0.0/20", "88.2.192.0/20"},
	},
	{
		ID: "ES-MAD-DELICIAS", ISP: "Telefónica de España (RIMA)",
		Name: "Central Delicias / Arganzuela (Paseo Delicias)", District: "Arganzuela / Delicias",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.3995, Lon: -3.6940, PrecisionKm: 1.2,
		CIDRs: []string{"2.139.8.0/22", "80.58.28.0/22", "83.35.32.0/20", "88.3.0.0/20"},
	},
	{
		ID: "ES-MAD-CARABANCHEL", ISP: "Telefónica de España (RIMA)",
		Name: "Central Buenavista / Carabanchel (General Ricardos)", District: "Carabanchel / Usera",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.3780, Lon: -3.7430, PrecisionKm: 1.4,
		CIDRs: []string{"2.139.0.0/22", "80.58.32.0/22", "83.35.64.0/20", "88.3.32.0/20"},
	},
	{
		ID: "ES-MAD-ALUCHE", ISP: "Telefónica de España (RIMA)",
		Name: "Central Aluche / Latina (Calle Maqueda)", District: "Latina / Aluche",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.3880, Lon: -3.7620, PrecisionKm: 1.4,
		CIDRs: []string{"2.139.4.0/22", "80.58.36.0/22", "83.35.96.0/20", "88.3.64.0/20"},
	},
	{
		ID: "ES-MAD-CONCEPCION", ISP: "Telefónica de España (RIMA)",
		Name: "Central Concepción / Ciudad Lineal (Arturo Soria)", District: "Ciudad Lineal / San Blas",
		City: "Madrid", Province: "Madrid", CountryCode: "ES",
		Lat: 40.4430, Lon: -3.6510, PrecisionKm: 1.3,
		CIDRs: []string{"2.139.12.0/22", "80.58.40.0/22", "83.35.128.0/20", "88.3.96.0/20"},
	},
	{
		ID: "ES-MAD-ALCOBENDAS", ISP: "Telefónica de España (RIMA)",
		Name: "Central Alcobendas / San Sebastián de los Reyes", District: "Alcobendas / Norte",
		City: "Alcobendas", Province: "Madrid", CountryCode: "ES",
		Lat: 40.5410, Lon: -3.6380, PrecisionKm: 1.5,
		CIDRs: []string{"2.139.36.0/22", "80.58.44.0/22", "83.36.0.0/20"},
	},
	{
		ID: "ES-MAD-GETAFE", ISP: "Telefónica de España (RIMA)",
		Name: "Central Getafe / Leganés Sur", District: "Getafe Centro",
		City: "Getafe", Province: "Madrid", CountryCode: "ES",
		Lat: 40.3060, Lon: -3.7310, PrecisionKm: 1.5,
		CIDRs: []string{"2.139.40.0/22", "80.58.48.0/22", "83.36.32.0/20"},
	},

	// ==================== BARCELONA CAPITAL Y METROPOLITANA ====================
	{
		ID: "ES-BCN-EIXAMPLE", ISP: "Telefónica de España (RIMA)",
		Name: "Central Eixample (Comte d'Urgell)", District: "L'Eixample",
		City: "Barcelona", Province: "Barcelona", CountryCode: "ES",
		Lat: 41.3850, Lon: 2.1550, PrecisionKm: 1.2,
		CIDRs: []string{"2.136.0.0/20", "80.58.64.0/20", "81.36.64.0/20", "83.40.0.0/19"},
	},
	{
		ID: "ES-BCN-GRACIA", ISP: "Telefónica de España (RIMA)",
		Name: "Central Gràcia (Travessera de Gràcia)", District: "Gràcia",
		City: "Barcelona", Province: "Barcelona", CountryCode: "ES",
		Lat: 41.4030, Lon: 2.1580, PrecisionKm: 1.1,
		CIDRs: []string{"2.136.16.0/20", "80.58.80.0/20", "83.40.32.0/19"},
	},
	{
		ID: "ES-BCN-POBLENOU", ISP: "Telefónica de España (RIMA)",
		Name: "Central Poblenou / Sant Martí (Rambla del Poblenou)", District: "Sant Martí / 22@",
		City: "Barcelona", Province: "Barcelona", CountryCode: "ES",
		Lat: 41.4010, Lon: 2.2020, PrecisionKm: 1.2,
		CIDRs: []string{"2.136.32.0/20", "80.58.96.0/20", "83.40.64.0/19"},
	},
	{
		ID: "ES-BCN-SANTS", ISP: "Telefónica de España (RIMA)",
		Name: "Central Sants / Montjuïc (Carrer de Sants)", District: "Sants-Montjuïc",
		City: "Barcelona", Province: "Barcelona", CountryCode: "ES",
		Lat: 41.3760, Lon: 2.1350, PrecisionKm: 1.3,
		CIDRs: []string{"2.136.48.0/20", "80.58.112.0/20", "83.40.96.0/19"},
	},

	// ==================== VALENCIA, SEVILLA, BILBAO, ZARAGOZA ====================
	{
		ID: "ES-VLC-ALAMEDA", ISP: "Telefónica de España (RIMA)",
		Name: "Central Alameda / Mestalla", District: "El Pla del Real / Mestalla",
		City: "Valencia", Province: "Valencia", CountryCode: "ES",
		Lat: 39.4750, Lon: -0.3640, PrecisionKm: 1.3,
		CIDRs: []string{"2.137.0.0/20", "80.58.128.0/20", "83.45.0.0/19"},
	},
	{
		ID: "ES-VLC-RUZAFACENTRO", ISP: "Telefónica de España (RIMA)",
		Name: "Central Ruzafa / Ensanche", District: "L'Eixample / Ruzafa",
		City: "Valencia", Province: "Valencia", CountryCode: "ES",
		Lat: 39.4610, Lon: -0.3740, PrecisionKm: 1.2,
		CIDRs: []string{"2.137.16.0/20", "80.58.144.0/20", "83.45.32.0/19"},
	},
	{
		ID: "ES-SEV-NERVION", ISP: "Telefónica de España (RIMA)",
		Name: "Central Nervión / Eduardo Dato", District: "Nervión",
		City: "Sevilla", Province: "Sevilla", CountryCode: "ES",
		Lat: 37.3820, Lon: -5.9730, PrecisionKm: 1.3,
		CIDRs: []string{"2.138.0.0/20", "80.58.160.0/20", "83.50.0.0/19"},
	},
	{
		ID: "ES-SEV-TRIANA", ISP: "Telefónica de España (RIMA)",
		Name: "Central Triana / Los Remedios (Calle San Jacinto)", District: "Triana / Los Remedios",
		City: "Sevilla", Province: "Sevilla", CountryCode: "ES",
		Lat: 37.3830, Lon: -6.0040, PrecisionKm: 1.3,
		CIDRs: []string{"2.138.16.0/20", "80.58.176.0/20", "83.50.32.0/19"},
	},
	{
		ID: "ES-BIO-ENSANCHE", ISP: "Telefónica de España (RIMA)",
		Name: "Central Bilbao Ensanche (Gran Vía Don Diego)", District: "Abando / Ensanche",
		City: "Bilbao", Province: "Bizkaia", CountryCode: "ES",
		Lat: 43.2620, Lon: -2.9340, PrecisionKm: 1.2,
		CIDRs: []string{"2.140.0.0/20", "80.58.192.0/20", "83.55.0.0/19"},
	},
	{
		ID: "ES-ZAZ-CENTRO", ISP: "Telefónica de España (RIMA)",
		Name: "Central Zaragoza Centro (Paseo Independencia)", District: "Centro",
		City: "Zaragoza", Province: "Zaragoza", CountryCode: "ES",
		Lat: 41.6500, Lon: -0.8830, PrecisionKm: 1.3,
		CIDRs: []string{"2.141.0.0/20", "80.58.208.0/20", "83.60.0.0/19"},
	},
}

// LookupCentralTelefonica checks if an IP belongs to a known telecom central office sub-allocation
func LookupCentralTelefonica(ipStr string) *CentralRecord {
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return nil
	}

	for i := range knownCentrales {
		c := &knownCentrales[i]
		for _, cidr := range c.CIDRs {
			_, subnet, err := net.ParseCIDR(cidr)
			if err == nil && subnet.Contains(parsedIP) {
				return c
			}
		}
	}

	return nil
}
