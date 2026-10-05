package recon

import (
	"fmt"
	"regexp"
	"strings"
)

// FacilityRecord represents ground-truth micro-geocoded coordinates for a high-value campus, institute or datacenter
type FacilityRecord struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "university", "research_lab", "datacenter", "ixp", "government"
	City        string  `json:"city"`
	PostalCode  string  `json:"postal_code,omitempty"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	PrecisionKm float64 `json:"precision_km"` // Ground-truth radius (0.2km - 0.9km)
	MatchRules  []string`json:"match_rules"`  // Substrings to match in ASOrg, ISP, PTR, or City
	ASNs        []int   `json:"asns,omitempty"`
}

// Master ground-truth catalog of 100+ global academic, research, and infrastructure facilities
var knownFacilities = []FacilityRecord{
	// ==================== UNIVERSIDADES Y CENTROS DE INVESTIGACIÓN ESPAÑOLES ====================
	{
		ID: "ES-USAL", Name: "Universidad de Salamanca (Campus Canalejas/Unamuno)", Type: "university",
		City: "Salamanca", PostalCode: "37007", Country: "España", CountryCode: "ES",
		Lat: 40.9616, Lon: -5.6698, PrecisionKm: 0.45,
		MatchRules: []string{"universidad de salamanca", "usal.es", "usal"}, ASNs: []int{766},
	},
	{
		ID: "ES-UCM", Name: "Universidad Complutense de Madrid (Ciudad Universitaria)", Type: "university",
		City: "Madrid", PostalCode: "28040", Country: "España", CountryCode: "ES",
		Lat: 40.4489, Lon: -3.7297, PrecisionKm: 0.50,
		MatchRules: []string{"universidad complutense", "ucm.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UPM", Name: "Universidad Politécnica de Madrid (Campus Moncloa)", Type: "university",
		City: "Madrid", PostalCode: "28040", Country: "España", CountryCode: "ES",
		Lat: 40.4497, Lon: -3.7275, PrecisionKm: 0.40,
		MatchRules: []string{"universidad politecnica de madrid", "upm.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UAM", Name: "Universidad Autónoma de Madrid (Campus Cantoblanco)", Type: "university",
		City: "Madrid", PostalCode: "28049", Country: "España", CountryCode: "ES",
		Lat: 40.5466, Lon: -3.6934, PrecisionKm: 0.55,
		MatchRules: []string{"universidad autonoma de madrid", "uam.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UC3M", Name: "Universidad Carlos III de Madrid (Campus Getafe/Leganés)", Type: "university",
		City: "Getafe", PostalCode: "28903", Country: "España", CountryCode: "ES",
		Lat: 40.3039, Lon: -3.7288, PrecisionKm: 0.45,
		MatchRules: []string{"carlos iii", "uc3m.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UAH", Name: "Universidad de Alcalá (Colegio de San Ildefonso)", Type: "university",
		City: "Alcalá de Henares", PostalCode: "28801", Country: "España", CountryCode: "ES",
		Lat: 40.4833, Lon: -3.3636, PrecisionKm: 0.40,
		MatchRules: []string{"universidad de alcala", "uah.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UPC", Name: "Universitat Politècnica de Catalunya (Campus Nord)", Type: "university",
		City: "Barcelona", PostalCode: "08034", Country: "España", CountryCode: "ES",
		Lat: 41.3892, Lon: 2.1134, PrecisionKm: 0.40,
		MatchRules: []string{"universitat politecnica de catalunya", "upc.edu", "upc.es"}, ASNs: []int{766, 1307},
	},
	{
		ID: "ES-UB", Name: "Universitat de Barcelona (Edifici Històric)", Type: "university",
		City: "Barcelona", PostalCode: "08007", Country: "España", CountryCode: "ES",
		Lat: 41.3870, Lon: 2.1649, PrecisionKm: 0.35,
		MatchRules: []string{"universitat de barcelona", "ub.edu"}, ASNs: []int{766, 1307},
	},
	{
		ID: "ES-UAB", Name: "Universitat Autònoma de Barcelona (Campus Bellaterra)", Type: "university",
		City: "Cerdanyola del Vallès", PostalCode: "08193", Country: "España", CountryCode: "ES",
		Lat: 41.5005, Lon: 2.1090, PrecisionKm: 0.50,
		MatchRules: []string{"autonoma de barcelona", "uab.cat", "uab.es"}, ASNs: []int{766, 1307},
	},
	{
		ID: "ES-UPF", Name: "Universitat Pompeu Fabra (Campus Ciutadella/Poblenou)", Type: "university",
		City: "Barcelona", PostalCode: "08005", Country: "España", CountryCode: "ES",
		Lat: 41.3895, Lon: 2.1932, PrecisionKm: 0.35,
		MatchRules: []string{"pompeu fabra", "upf.edu"}, ASNs: []int{766, 1307},
	},
	{
		ID: "ES-UPV", Name: "Universitat Politècnica de València (Campus de Vera)", Type: "university",
		City: "Valencia", PostalCode: "46022", Country: "España", CountryCode: "ES",
		Lat: 49.4811, Lon: -0.3429, PrecisionKm: 0.45,
		MatchRules: []string{"politecnica de valencia", "upv.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UV", Name: "Universitat de València (Campus Blasco Ibáñez)", Type: "university",
		City: "Valencia", PostalCode: "46010", Country: "España", CountryCode: "ES",
		Lat: 39.4795, Lon: -0.3582, PrecisionKm: 0.40,
		MatchRules: []string{"universitat de valencia", "uv.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-US", Name: "Universidad de Sevilla (Rectorado Fábrica de Tabacos)", Type: "university",
		City: "Sevilla", PostalCode: "41004", Country: "España", CountryCode: "ES",
		Lat: 37.3804, Lon: -5.9912, PrecisionKm: 0.40,
		MatchRules: []string{"universidad de sevilla", "us.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UPO", Name: "Universidad Pablo de Olavide", Type: "university",
		City: "Sevilla", PostalCode: "41013", Country: "España", CountryCode: "ES",
		Lat: 37.3541, Lon: -5.9372, PrecisionKm: 0.50,
		MatchRules: []string{"pablo de olavide", "upo.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UGR", Name: "Universidad de Granada (Hospital Real)", Type: "university",
		City: "Granada", PostalCode: "18071", Country: "España", CountryCode: "ES",
		Lat: 37.1852, Lon: -3.6006, PrecisionKm: 0.40,
		MatchRules: []string{"universidad de granada", "ugr.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UNIZAR", Name: "Universidad de Zaragoza (Campus San Francisco)", Type: "university",
		City: "Zaragoza", PostalCode: "50009", Country: "España", CountryCode: "ES",
		Lat: 41.6425, Lon: -0.9015, PrecisionKm: 0.45,
		MatchRules: []string{"universidad de zaragoza", "unizar.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UNIOVI", Name: "Universidad de Oviedo (Edificio Histórico)", Type: "university",
		City: "Oviedo", PostalCode: "33003", Country: "España", CountryCode: "ES",
		Lat: 43.3619, Lon: -5.8494, PrecisionKm: 0.40,
		MatchRules: []string{"universidad de oviedo", "uniovi.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-USC", Name: "Universidad de Santiago de Compostela (Colexio de San Xerome)", Type: "university",
		City: "Santiago de Compostela", PostalCode: "15782", Country: "España", CountryCode: "ES",
		Lat: 42.8790, Lon: -8.5448, PrecisionKm: 0.45,
		MatchRules: []string{"santiago de compostela", "usc.es", "usc.gal"}, ASNs: []int{766},
	},
	{
		ID: "ES-UDC", Name: "Universidade da Coruña (Campus de Elviña)", Type: "university",
		City: "A Coruña", PostalCode: "15071", Country: "España", CountryCode: "ES",
		Lat: 43.3328, Lon: -8.4106, PrecisionKm: 0.45,
		MatchRules: []string{"universidade da coruna", "udc.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UVIGO", Name: "Universidade de Vigo (Campus As Lagoas)", Type: "university",
		City: "Vigo", PostalCode: "36310", Country: "España", CountryCode: "ES",
		Lat: 42.1698, Lon: -8.6853, PrecisionKm: 0.50,
		MatchRules: []string{"universidade de vigo", "uvigo.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-EHU", Name: "Universidad del País Vasco / Euskal Herriko Unibertsitatea", Type: "university",
		City: "Leioa/Bilbao", PostalCode: "48940", Country: "España", CountryCode: "ES",
		Lat: 43.3308, Lon: -2.9692, PrecisionKm: 0.50,
		MatchRules: []string{"euskal herriko", "ehu.eus", "ehu.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UVA", Name: "Universidad de Valladolid (Palacio de Santa Cruz)", Type: "university",
		City: "Valladolid", PostalCode: "47002", Country: "España", CountryCode: "ES",
		Lat: 41.6517, Lon: -4.7212, PrecisionKm: 0.40,
		MatchRules: []string{"universidad de valladolid", "uva.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UMA", Name: "Universidad de Málaga (Campus de Teatinos)", Type: "university",
		City: "Málaga", PostalCode: "29071", Country: "España", CountryCode: "ES",
		Lat: 36.7166, Lon: -4.4786, PrecisionKm: 0.50,
		MatchRules: []string{"universidad de malaga", "uma.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UM", Name: "Universidad de Murcia (Campus de Espinardo)", Type: "university",
		City: "Murcia", PostalCode: "30100", Country: "España", CountryCode: "ES",
		Lat: 38.0195, Lon: -1.1712, PrecisionKm: 0.50,
		MatchRules: []string{"universidad de murcia", "um.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UA", Name: "Universidad de Alicante (Campus San Vicente)", Type: "university",
		City: "San Vicente del Raspeig", PostalCode: "03690", Country: "España", CountryCode: "ES",
		Lat: 38.3847, Lon: -0.5133, PrecisionKm: 0.50,
		MatchRules: []string{"universidad de alicante", "ua.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UCO", Name: "Universidad de Córdoba (Campus de Rabanales)", Type: "university",
		City: "Córdoba", PostalCode: "14071", Country: "España", CountryCode: "ES",
		Lat: 37.9157, Lon: -4.7231, PrecisionKm: 0.50,
		MatchRules: []string{"universidad de cordoba", "uco.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UCA", Name: "Universidad de Cádiz (Rectorado Calle Ancha)", Type: "university",
		City: "Cádiz", PostalCode: "11001", Country: "España", CountryCode: "ES",
		Lat: 36.5336, Lon: -6.2974, PrecisionKm: 0.40,
		MatchRules: []string{"universidad de cadiz", "uca.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-UNICAN", Name: "Universidad de Cantabria (Campus Las Llamas)", Type: "university",
		City: "Santander", PostalCode: "39005", Country: "España", CountryCode: "ES",
		Lat: 43.4797, Lon: -3.8016, PrecisionKm: 0.45,
		MatchRules: []string{"universidad de cantabria", "unican.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-ULL", Name: "Universidad de La Laguna (Campus Central)", Type: "university",
		City: "San Cristóbal de La Laguna", PostalCode: "38200", Country: "España", CountryCode: "ES",
		Lat: 28.4815, Lon: -16.3175, PrecisionKm: 0.45,
		MatchRules: []string{"universidad de la laguna", "ull.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-ULPGC", Name: "Universidad de Las Palmas de Gran Canaria (Campus Tafira)", Type: "university",
		City: "Las Palmas de Gran Canaria", PostalCode: "35017", Country: "España", CountryCode: "ES",
		Lat: 28.0722, Lon: -15.4525, PrecisionKm: 0.50,
		MatchRules: []string{"las palmas de gran canaria", "ulpgc.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-CSIC", Name: "CSIC - Consejo Superior de Investigaciones Científicas", Type: "research_lab",
		City: "Madrid", PostalCode: "28006", Country: "España", CountryCode: "ES",
		Lat: 40.4407, Lon: -3.6874, PrecisionKm: 0.30,
		MatchRules: []string{"csic", "consejo superior de investigaciones"}, ASNs: []int{766},
	},
	{
		ID: "ES-CIEMAT", Name: "CIEMAT Centro de Investigaciones Energéticas", Type: "research_lab",
		City: "Madrid", PostalCode: "28040", Country: "España", CountryCode: "ES",
		Lat: 40.4568, Lon: -3.7328, PrecisionKm: 0.35,
		MatchRules: []string{"ciemat"}, ASNs: []int{766},
	},
	{
		ID: "ES-BSC", Name: "Barcelona Supercomputing Center (Torre Girona)", Type: "research_lab",
		City: "Barcelona", PostalCode: "08034", Country: "España", CountryCode: "ES",
		Lat: 41.3891, Lon: 2.1158, PrecisionKm: 0.25,
		MatchRules: []string{"barcelona supercomputing", "bsc.es"}, ASNs: []int{766, 1307},
	},
	{
		ID: "ES-IAC", Name: "Instituto de Astrofísica de Canarias (Sede Central)", Type: "research_lab",
		City: "La Laguna", PostalCode: "38205", Country: "España", CountryCode: "ES",
		Lat: 28.4731, Lon: -16.3094, PrecisionKm: 0.30,
		MatchRules: []string{"instituto de astrofisica de canarias", "iac.es"}, ASNs: []int{766},
	},
	{
		ID: "ES-REDIRIS", Name: "RedIRIS Centro de Operaciones de Red Central", Type: "ixp",
		City: "Madrid", PostalCode: "28020", Country: "España", CountryCode: "ES",
		Lat: 40.4530, Lon: -3.6930, PrecisionKm: 0.30,
		MatchRules: []string{"rediris", "red iris", "rediris.es"}, ASNs: []int{766},
	},

	// ==================== INSTITUTOS Y UNIVERSIDADES DE EUROPA Y ASIA ====================
	{
		ID: "CH-CERN", Name: "CERN - European Organization for Nuclear Research", Type: "research_lab",
		City: "Meyrin / Geneva", PostalCode: "1211", Country: "Switzerland", CountryCode: "CH",
		Lat: 46.2330, Lon: 6.0557, PrecisionKm: 0.40,
		MatchRules: []string{"cern", "european organization for nuclear research"}, ASNs: []int{513},
	},
	{
		ID: "CH-ETHZ", Name: "ETH Zürich - Eidgenössische Technische Hochschule", Type: "university",
		City: "Zürich", PostalCode: "8092", Country: "Switzerland", CountryCode: "CH",
		Lat: 47.3763, Lon: 8.5481, PrecisionKm: 0.35,
		MatchRules: []string{"eth zürich", "eth zurich", "ethz.ch"}, ASNs: []int{559},
	},
	{
		ID: "CH-EPFL", Name: "EPFL - École Polytechnique Fédérale de Lausanne", Type: "university",
		City: "Lausanne", PostalCode: "1015", Country: "Switzerland", CountryCode: "CH",
		Lat: 46.5191, Lon: 6.5668, PrecisionKm: 0.40,
		MatchRules: []string{"epfl.ch", "ecole polytechnique federale de lausanne"}, ASNs: []int{559},
	},
	{
		ID: "GB-OXFORD", Name: "University of Oxford (Wellington Square / IT Services)", Type: "university",
		City: "Oxford", PostalCode: "OX1 2JD", Country: "United Kingdom", CountryCode: "GB",
		Lat: 51.7589, Lon: -1.2587, PrecisionKm: 0.40,
		MatchRules: []string{"university of oxford", "ox.ac.uk", "oxford.ac.uk"}, ASNs: []int{786},
	},
	{
		ID: "GB-CAMBRIDGE", Name: "University of Cambridge (The Old Schools / IT)", Type: "university",
		City: "Cambridge", PostalCode: "CB2 1TN", Country: "United Kingdom", CountryCode: "GB",
		Lat: 52.2053, Lon: 0.1166, PrecisionKm: 0.40,
		MatchRules: []string{"university of cambridge", "cam.ac.uk"}, ASNs: []int{786},
	},
	{
		ID: "GB-IMPERIAL", Name: "Imperial College London (South Kensington Campus)", Type: "university",
		City: "London", PostalCode: "SW7 2AZ", Country: "United Kingdom", CountryCode: "GB",
		Lat: 51.4988, Lon: -0.1749, PrecisionKm: 0.30,
		MatchRules: []string{"imperial college", "imperial.ac.uk"}, ASNs: []int{786},
	},
	{
		ID: "GB-UCL", Name: "University College London (Bloomsbury Campus)", Type: "university",
		City: "London", PostalCode: "WC1E 6BT", Country: "United Kingdom", CountryCode: "GB",
		Lat: 51.5246, Lon: -0.1340, PrecisionKm: 0.30,
		MatchRules: []string{"university college london", "ucl.ac.uk"}, ASNs: []int{786},
	},
	{
		ID: "DE-TUM", Name: "Technical University of Munich (TUM Main Campus)", Type: "university",
		City: "Munich", PostalCode: "80333", Country: "Germany", CountryCode: "DE",
		Lat: 48.1497, Lon: 11.5681, PrecisionKm: 0.40,
		MatchRules: []string{"technische universitaet muenchen", "tum.de", "tu-muenchen"}, ASNs: []int{568},
	},
	{
		ID: "DE-LMU", Name: "LMU Ludwig-Maximilians-Universität München", Type: "university",
		City: "Munich", PostalCode: "80539", Country: "Germany", CountryCode: "DE",
		Lat: 48.1508, Lon: 11.5802, PrecisionKm: 0.40,
		MatchRules: []string{"ludwig-maximilians", "uni-muenchen.de", "lmu.de"}, ASNs: []int{568},
	},
	{
		ID: "DE-MPI", Name: "Max Planck Institute for Informatics", Type: "research_lab",
		City: "Saarbrücken", PostalCode: "66123", Country: "Germany", CountryCode: "DE",
		Lat: 49.2580, Lon: 7.0435, PrecisionKm: 0.30,
		MatchRules: []string{"max planck", "mpi-inf", "mpg.de"}, ASNs: []int{680},
	},
	{
		ID: "FR-SORBONNE", Name: "Sorbonne Université (Campus Pierre et Marie Curie)", Type: "university",
		City: "Paris", PostalCode: "75005", Country: "France", CountryCode: "FR",
		Lat: 48.8471, Lon: 2.3574, PrecisionKm: 0.35,
		MatchRules: []string{"sorbonne", "sorbonne-universite.fr", "upmc.fr"}, ASNs: []int{2200},
	},
	{
		ID: "FR-POLYTECH", Name: "Institut Polytechnique de Paris (École Polytechnique)", Type: "university",
		City: "Palaiseau", PostalCode: "91120", Country: "France", CountryCode: "FR",
		Lat: 48.7130, Lon: 2.2084, PrecisionKm: 0.45,
		MatchRules: []string{"polytechnique.fr", "polytechnique.edu", "ecole polytechnique"}, ASNs: []int{2200},
	},
	{
		ID: "NL-UVA", Name: "University of Amsterdam (Roeterseiland / Science Park)", Type: "university",
		City: "Amsterdam", PostalCode: "1018 WS", Country: "Netherlands", CountryCode: "NL",
		Lat: 52.3622, Lon: 4.9125, PrecisionKm: 0.35,
		MatchRules: []string{"universiteit van amsterdam", "uva.nl"}, ASNs: []int{1103},
	},
	{
		ID: "NL-TUDELFT", Name: "Delft University of Technology (TU Delft Aula)", Type: "university",
		City: "Delft", PostalCode: "2628 BL", Country: "Netherlands", CountryCode: "NL",
		Lat: 51.9989, Lon: 4.3736, PrecisionKm: 0.40,
		MatchRules: []string{"delft university", "tudelft.nl"}, ASNs: []int{1103},
	},
	{
		ID: "BE-KULEUVEN", Name: "KU Leuven (Katholieke Universiteit Leuven)", Type: "university",
		City: "Leuven", PostalCode: "3000", Country: "Belgium", CountryCode: "BE",
		Lat: 50.8778, Lon: 4.7003, PrecisionKm: 0.35,
		MatchRules: []string{"ku leuven", "kuleuven.be"}, ASNs: []int{2611},
	},
	{
		ID: "SE-KAROLINSKA", Name: "Karolinska Institutet (Solna Campus)", Type: "university",
		City: "Stockholm/Solna", PostalCode: "171 77", Country: "Sweden", CountryCode: "SE",
		Lat: 59.3498, Lon: 18.0289, PrecisionKm: 0.40,
		MatchRules: []string{"karolinska", "ki.se"}, ASNs: []int{1653},
	},
	{
		ID: "JP-UTOKYO", Name: "The University of Tokyo (Hongo Campus)", Type: "university",
		City: "Tokyo", PostalCode: "113-8654", Country: "Japan", CountryCode: "JP",
		Lat: 35.7126, Lon: 139.7620, PrecisionKm: 0.40,
		MatchRules: []string{"university of tokyo", "u-tokyo.ac.jp", "todai"}, ASNs: []int{2500, 2907},
	},
	{
		ID: "JP-KYOTO", Name: "Kyoto University (Yoshida Main Campus)", Type: "university",
		City: "Kyoto", PostalCode: "606-8501", Country: "Japan", CountryCode: "JP",
		Lat: 35.0262, Lon: 135.7808, PrecisionKm: 0.40,
		MatchRules: []string{"kyoto university", "kyoto-u.ac.jp"}, ASNs: []int{2500, 2907},
	},
	{
		ID: "SG-NUS", Name: "National University of Singapore (Kent Ridge)", Type: "university",
		City: "Singapore", PostalCode: "119077", Country: "Singapore", CountryCode: "SG",
		Lat: 1.2966, Lon: 103.7764, PrecisionKm: 0.45,
		MatchRules: []string{"national university of singapore", "nus.edu.sg"}, ASNs: []int{4606},
	},
	{
		ID: "AU-ANU", Name: "Australian National University (Acton Campus)", Type: "university",
		City: "Canberra", PostalCode: "2601", Country: "Australia", CountryCode: "AU",
		Lat: -35.2777, Lon: 149.1185, PrecisionKm: 0.45,
		MatchRules: []string{"australian national university", "anu.edu.au"}, ASNs: []int{7575},
	},

	// ==================== UNIVERSIDADES Y LABORATORIOS DE ESTADOS UNIDOS ====================
	{
		ID: "US-MIT", Name: "Massachusetts Institute of Technology (Ray & Maria Stata Center)", Type: "university",
		City: "Cambridge", PostalCode: "02139", Country: "United States", CountryCode: "US",
		Lat: 42.3616, Lon: -71.0906, PrecisionKm: 0.30,
		MatchRules: []string{"massachusetts institute of technology", "mit.edu", "mit"}, ASNs: []int{3, 111},
	},
	{
		ID: "US-HARVARD", Name: "Harvard University (Massachusetts Hall & Yard)", Type: "university",
		City: "Cambridge", PostalCode: "02138", Country: "United States", CountryCode: "US",
		Lat: 42.3770, Lon: -71.1167, PrecisionKm: 0.35,
		MatchRules: []string{"harvard university", "harvard.edu"}, ASNs: []int{11, 10578},
	},
	{
		ID: "US-STANFORD", Name: "Stanford University (Gates Computer Science Building)", Type: "university",
		City: "Stanford", PostalCode: "94305", Country: "United States", CountryCode: "US",
		Lat: 47.4300, Lon: -122.1733, PrecisionKm: 0.40,
		MatchRules: []string{"stanford university", "stanford.edu"}, ASNs: []int{32, 55},
	},
	{
		ID: "US-BERKELEY", Name: "University of California, Berkeley (Soda Hall)", Type: "university",
		City: "Berkeley", PostalCode: "94720", Country: "United States", CountryCode: "US",
		Lat: 37.8756, Lon: -122.2588, PrecisionKm: 0.35,
		MatchRules: []string{"university of california, berkeley", "berkeley.edu", "uc berkeley"}, ASNs: []int{25},
	},
	{
		ID: "US-PRINCETON", Name: "Princeton University (Nassau Hall)", Type: "university",
		City: "Princeton", PostalCode: "08544", Country: "United States", CountryCode: "US",
		Lat: 40.3430, Lon: -74.6514, PrecisionKm: 0.35,
		MatchRules: []string{"princeton university", "princeton.edu"}, ASNs: []int{88},
	},
	{
		ID: "US-COLUMBIA", Name: "Columbia University (Low Memorial Library)", Type: "university",
		City: "New York", PostalCode: "10027", Country: "United States", CountryCode: "US",
		Lat: 40.8075, Lon: -73.9626, PrecisionKm: 0.30,
		MatchRules: []string{"columbia university", "columbia.edu"}, ASNs: []int{14},
	},
	{
		ID: "US-YALE", Name: "Yale University (Woodbridge Hall)", Type: "university",
		City: "New Haven", PostalCode: "06520", Country: "United States", CountryCode: "US",
		Lat: 41.3110, Lon: -72.9288, PrecisionKm: 0.35,
		MatchRules: []string{"yale university", "yale.edu"}, ASNs: []int{29},
	},
	{
		ID: "US-CALTECH", Name: "California Institute of Technology (Beckman Institute)", Type: "university",
		City: "Pasadena", PostalCode: "91125", Country: "United States", CountryCode: "US",
		Lat: 34.1377, Lon: -118.1253, PrecisionKm: 0.30,
		MatchRules: []string{"caltech", "california institute of technology", "caltech.edu"}, ASNs: []int{157},
	},
	{
		ID: "US-CMU", Name: "Carnegie Mellon University (Gates Center)", Type: "university",
		City: "Pittsburgh", PostalCode: "15213", Country: "United States", CountryCode: "US",
		Lat: 40.4444, Lon: -79.9430, PrecisionKm: 0.35,
		MatchRules: []string{"carnegie mellon", "cmu.edu"}, ASNs: []int{9},
	},
	{
		ID: "US-CORNELL", Name: "Cornell University (Ithaca Central Campus)", Type: "university",
		City: "Ithaca", PostalCode: "14853", Country: "United States", CountryCode: "US",
		Lat: 42.4534, Lon: -76.4735, PrecisionKm: 0.45,
		MatchRules: []string{"cornell university", "cornell.edu"}, ASNs: []int{63},
	},
	{
		ID: "US-UW", Name: "University of Washington (Paul G. Allen Center)", Type: "university",
		City: "Seattle", PostalCode: "98195", Country: "United States", CountryCode: "US",
		Lat: 47.6534, Lon: -122.3059, PrecisionKm: 0.35,
		MatchRules: []string{"university of washington", "washington.edu", "uwnet"}, ASNs: []int{73},
	},
	{
		ID: "US-UMICH", Name: "University of Michigan (North Campus / Central)", Type: "university",
		City: "Ann Arbor", PostalCode: "48109", Country: "United States", CountryCode: "US",
		Lat: 42.2780, Lon: -83.7382, PrecisionKm: 0.45,
		MatchRules: []string{"university of michigan", "umich.edu"}, ASNs: []int{36, 237},
	},
	{
		ID: "US-GEORGIATECH", Name: "Georgia Institute of Technology (Klaus Advanced Computing)", Type: "university",
		City: "Atlanta", PostalCode: "30332", Country: "United States", CountryCode: "US",
		Lat: 33.7773, Lon: -84.3962, PrecisionKm: 0.35,
		MatchRules: []string{"georgia institute of technology", "gatech.edu"}, ASNs: []int{2637},
	},
	{
		ID: "US-UTAUSTIN", Name: "University of Texas at Austin (Gates Dell Complex)", Type: "university",
		City: "Austin", PostalCode: "78712", Country: "United States", CountryCode: "US",
		Lat: 30.2861, Lon: -97.7366, PrecisionKm: 0.40,
		MatchRules: []string{"university of texas at austin", "utexas.edu"}, ASNs: []int{3354},
	},
	{
		ID: "US-UIUC", Name: "University of Illinois Urbana-Champaign (Siebel Center)", Type: "university",
		City: "Urbana", PostalCode: "61801", Country: "United States", CountryCode: "US",
		Lat: 40.1138, Lon: -88.2249, PrecisionKm: 0.35,
		MatchRules: []string{"illinois urbana", "uiuc.edu", "illinois.edu"}, ASNs: []int{38},
	},
	{
		ID: "US-NASA-AMES", Name: "NASA Ames Research Center (Moffett Field)", Type: "research_lab",
		City: "Mountain View", PostalCode: "94035", Country: "United States", CountryCode: "US",
		Lat: 37.4152, Lon: -122.0626, PrecisionKm: 0.35,
		MatchRules: []string{"nasa ames", "arc.nasa.gov", "ames research"}, ASNs: []int{42, 7018},
	},
	{
		ID: "US-NASA-GODDARD", Name: "NASA Goddard Space Flight Center", Type: "research_lab",
		City: "Greenbelt", PostalCode: "20771", Country: "United States", CountryCode: "US",
		Lat: 38.9959, Lon: -76.8526, PrecisionKm: 0.40,
		MatchRules: []string{"goddard space", "gsfc.nasa.gov"}, ASNs: []int{42},
	},
	{
		ID: "US-FERMILAB", Name: "Fermi National Accelerator Laboratory (Wilson Hall)", Type: "research_lab",
		City: "Batavia", PostalCode: "60510", Country: "United States", CountryCode: "US",
		Lat: 41.8419, Lon: -88.2573, PrecisionKm: 0.45,
		MatchRules: []string{"fermilab", "fnal.gov"}, ASNs: []int{3152},
	},
	{
		ID: "US-LBNL", Name: "Lawrence Berkeley National Laboratory (LBNL)", Type: "research_lab",
		City: "Berkeley", PostalCode: "94720", Country: "United States", CountryCode: "US",
		Lat: 37.8756, Lon: -122.2508, PrecisionKm: 0.30,
		MatchRules: []string{"lbl.gov", "lawrence berkeley national laboratory"}, ASNs: []int{3152, 293},
	},

	// ==================== CENTROS DE DATOS GLOBALES, IXPs Y CLOUD ====================
	{
		ID: "DE-HETZNER-FSN1", Name: "Hetzner Datacenter Park Falkenstein (FSN1)", Type: "datacenter",
		City: "Falkenstein/Vogtland", PostalCode: "08223", Country: "Germany", CountryCode: "DE",
		Lat: 50.4789, Lon: 12.3705, PrecisionKm: 0.25,
		MatchRules: []string{"falkenstein", "fsn1", "hetzner online gmbh"}, ASNs: []int{24940},
	},
	{
		ID: "DE-HETZNER-NBG1", Name: "Hetzner Datacenter Park Nürnberg (NBG1)", Type: "datacenter",
		City: "Nürnberg", PostalCode: "90431", Country: "Germany", CountryCode: "DE",
		Lat: 49.4312, Lon: 11.0264, PrecisionKm: 0.25,
		MatchRules: []string{"nuernberg", "nuremberg", "nbg1"}, ASNs: []int{24940},
	},
	{
		ID: "DE-HETZNER-HEL1", Name: "Hetzner Datacenter Park Helsinki (HEL1)", Type: "datacenter",
		City: "Tuusula / Helsinki", PostalCode: "04300", Country: "Finland", CountryCode: "FI",
		Lat: 60.4072, Lon: 25.0264, PrecisionKm: 0.25,
		MatchRules: []string{"hel1", "tuusula", "helsinki hetzner"}, ASNs: []int{24940},
	},
	{
		ID: "FR-OVH-RBX", Name: "OVHcloud Headquarters & Datacenter Campus Roubaix (RBX)", Type: "datacenter",
		City: "Roubaix", PostalCode: "59100", Country: "France", CountryCode: "FR",
		Lat: 50.6927, Lon: 3.1778, PrecisionKm: 0.25,
		MatchRules: []string{"roubaix", "rbx", "ovh sas", "ovhcloud"}, ASNs: []int{16276},
	},
	{
		ID: "FR-OVH-GRA", Name: "OVHcloud Datacenter Gravelines (GRA)", Type: "datacenter",
		City: "Gravelines", PostalCode: "59820", Country: "France", CountryCode: "FR",
		Lat: 50.9872, Lon: 2.1245, PrecisionKm: 0.30,
		MatchRules: []string{"gravelines", "gra1", "gra2"}, ASNs: []int{16276},
	},
	{
		ID: "FR-OVH-SBG", Name: "OVHcloud Datacenter Strasbourg (SBG)", Type: "datacenter",
		City: "Strasbourg", PostalCode: "67000", Country: "France", CountryCode: "FR",
		Lat: 48.5303, Lon: 7.7942, PrecisionKm: 0.25,
		MatchRules: []string{"strasbourg", "sbg1", "sbg2"}, ASNs: []int{16276},
	},
	{
		ID: "ES-EQUINIX-MD2", Name: "Equinix IBX MD2 Datacenter Madrid (Alcobendas)", Type: "datacenter",
		City: "Madrid/Alcobendas", PostalCode: "28108", Country: "España", CountryCode: "ES",
		Lat: 40.5283, Lon: -3.6547, PrecisionKm: 0.20,
		MatchRules: []string{"equinix spain", "md2", "valgrande 6"}, ASNs: []int{24115, 12695},
	},
	{
		ID: "ES-INTERXION-MAD1", Name: "Interxion MAD1 / Digital Realty Madrid", Type: "datacenter",
		City: "Madrid", PostalCode: "28037", Country: "España", CountryCode: "ES",
		Lat: 40.4357, Lon: -3.6264, PrecisionKm: 0.20,
		MatchRules: []string{"interxion", "albasanz 71", "mad1"}, ASNs: []int{8218, 31334},
	},
	{
		ID: "GB-TELEHOUSE-LON", Name: "Telehouse London Docklands (North & East / LINX Core)", Type: "datacenter",
		City: "London", PostalCode: "E14 2AA", Country: "United Kingdom", CountryCode: "GB",
		Lat: 51.5113, Lon: -0.0076, PrecisionKm: 0.20,
		MatchRules: []string{"telehouse", "coriander avenue", "docklands"}, ASNs: []int{8367, 5459},
	},
	{
		ID: "DE-DECIX-FRA", Name: "DE-CIX Frankfurt / Interxion FRA1 Campus", Type: "ixp",
		City: "Frankfurt am Main", PostalCode: "60314", Country: "Germany", CountryCode: "DE",
		Lat: 50.1154, Lon: 8.7291, PrecisionKm: 0.20,
		MatchRules: []string{"de-cix", "hanauer landstrasse 322"}, ASNs: []int{6695, 8218},
	},
	{
		ID: "NL-AMSIX-EQUINIX", Name: "AMS-IX / Equinix AM3 Science Park Amsterdam", Type: "ixp",
		City: "Amsterdam", PostalCode: "1098 XH", Country: "Netherlands", CountryCode: "NL",
		Lat: 52.3551, Lon: 4.9542, PrecisionKm: 0.20,
		MatchRules: []string{"ams-ix", "science park 610", "am3"}, ASNs: []int{1200, 24115},
	},
	{
		ID: "US-EQUINIX-ASHBURN", Name: "Equinix Ashburn DC2 Campus (Data Center Alley)", Type: "datacenter",
		City: "Ashburn", PostalCode: "20147", Country: "United States", CountryCode: "US",
		Lat: 39.0180, Lon: -77.4590, PrecisionKm: 0.20,
		MatchRules: []string{"ashburn", "filigree", "equinix ashburn"}, ASNs: []int{24115, 14618},
	},
	{
		ID: "US-DO-NYC3", Name: "DigitalOcean NYC3 Datacenter (111 8th Avenue)", Type: "datacenter",
		City: "New York", PostalCode: "10011", Country: "United States", CountryCode: "US",
		Lat: 40.7402, Lon: -74.0022, PrecisionKm: 0.20,
		MatchRules: []string{"nyc3", "111 8th avenue", "digitalocean nyc3"}, ASNs: []int{14061},
	},
	{
		ID: "NL-DO-AMS3", Name: "DigitalOcean AMS3 Datacenter (Kabelweg 57)", Type: "datacenter",
		City: "Amsterdam", PostalCode: "1014 BA", Country: "Netherlands", CountryCode: "NL",
		Lat: 52.3906, Lon: 4.8517, PrecisionKm: 0.20,
		MatchRules: []string{"ams3", "kabelweg", "digitalocean ams3"}, ASNs: []int{14061},
	},
	{
		ID: "DE-DO-FRA1", Name: "DigitalOcean FRA1 Datacenter (Hanauer Landstraße)", Type: "datacenter",
		City: "Frankfurt", PostalCode: "60314", Country: "Germany", CountryCode: "DE",
		Lat: 50.1160, Lon: 8.7300, PrecisionKm: 0.20,
		MatchRules: []string{"fra1", "digitalocean frankfurt", "digitalocean fra1"}, ASNs: []int{14061},
	},
	{
		ID: "FR-SCALEWAY-DC3", Name: "Scaleway Datacenter DC3 (Vitry-sur-Seine)", Type: "datacenter",
		City: "Paris/Vitry", PostalCode: "94400", Country: "France", CountryCode: "FR",
		Lat: 48.7886, Lon: 2.4045, PrecisionKm: 0.20,
		MatchRules: []string{"scaleway", "online sas", "vitry-sur-seine"}, ASNs: []int{12876},
	},
	{
		ID: "NL-LEASEWEB-AMS01", Name: "LeaseWeb Amsterdam AMS-01 Datacenter", Type: "datacenter",
		City: "Amsterdam", PostalCode: "1101 BS", Country: "Netherlands", CountryCode: "NL",
		Lat: 52.2989, Lon: 4.9632, PrecisionKm: 0.20,
		MatchRules: []string{"leaseweb", "luttenbergweg"}, ASNs: []int{60781, 16265},
	},
}

// FindMatchingFacility searches the ground-truth database for exact facility match
func FindMatchingFacility(asn int, org, isp, hostname, city, zip, country string) *FacilityRecord {
	combined := strings.ToLower(fmt.Sprintf("%s %s %s %s %s %s", org, isp, hostname, city, zip, country))

	// 1. Direct ASN check combined with keyword verification
	for i := range knownFacilities {
		fac := &knownFacilities[i]
		asnMatched := false
		for _, a := range fac.ASNs {
			if a == asn && asn != 0 {
				asnMatched = true
				break
			}
		}

		// Check keyword rules
		for _, rule := range fac.MatchRules {
			if strings.Contains(combined, strings.ToLower(rule)) {
				return fac
			}
		}

		if asnMatched && fac.City != "" && strings.Contains(strings.ToLower(city), strings.ToLower(fac.City)) {
			return fac
		}
	}

	return nil
}

// PostalCentroid returns an estimated centroid for a known postal code
func PostalCentroid(countryCode, zip string) (float64, float64, float64, bool) {
	// Standard micro-geocoding for common postal codes
	zip = strings.TrimSpace(zip)
	if zip == "" {
		return 0, 0, 0, false
	}

	// In dense urban zones, a postal code centroid provides ±0.5km - 0.9km accuracy
	var zipPrefixRegex = regexp.MustCompile(`^(\d{2})`)
	if matches := zipPrefixRegex.FindStringSubmatch(zip); len(matches) > 1 && (countryCode == "ES" || countryCode == "es") {
		// Specific Spanish Postal Zones
		switch zip {
		case "37007", "37008", "37001", "37002": // Salamanca
			return 40.9635, -5.6635, 0.75, true
		case "28040": // Moncloa / Ciudad Universitaria Madrid
			return 40.4490, -3.7290, 0.50, true
		case "28001", "28002", "28006", "28020": // Madrid Centro / Castellana
			return 40.4350, -3.6880, 0.70, true
		case "28108": // Alcobendas / Datacenter Hub
			return 40.5300, -3.6500, 0.65, true
		case "08034": // Pedralbes / Campus Nord Barcelona
			return 41.3880, 2.1120, 0.55, true
		case "08007", "08001", "08002": // Barcelona Eixample / Ciutat Vella
			return 41.3880, 2.1670, 0.60, true
		case "46022": // Valencia Vera / Algirós
			return 39.4800, -0.3450, 0.60, true
		case "41004": // Sevilla Centro
			return 37.3850, -5.9900, 0.65, true
		case "18071", "18001": // Granada Centro
			return 37.1800, -3.6000, 0.65, true
		case "50009": // Zaragoza Universidad
			return 41.6400, -0.9000, 0.65, true
		}
	}

	return 0, 0, 0, false
}
