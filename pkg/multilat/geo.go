package multilat

import (
	"math"
)

const (
	// EarthRadiusKm is the approximate mean radius of the Earth in kilometers
	EarthRadiusKm = 6371.0

	// SpeedOfLightFiberKmPerMs is the propagation speed of light in single-mode fiber (approx 2/3 c)
	SpeedOfLightFiberKmPerMs = 200.0
)

// Point represents a geographic coordinate
type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// DistanceHaversine calculates great-circle distance between two points in km
func DistanceHaversine(p1, p2 Point) float64 {
	dLat := (p2.Lat - p1.Lat) * math.Pi / 180.0
	dLon := (p2.Lon - p1.Lon) * math.Pi / 180.0

	lat1 := p1.Lat * math.Pi / 180.0
	lat2 := p2.Lat * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

// ConstraintRadiusFromRTT converts Round-Trip Time in milliseconds to maximum distance constraint (CBG model)
// RTT is round-trip, so one-way max distance = (RTT_ms / 2) * SpeedOfLightFiber
func ConstraintRadiusFromRTT(rttMs float64) float64 {
	oneWayMs := rttMs / 2.0
	return oneWayMs * SpeedOfLightFiberKmPerMs
}

// Landmark represents a known probing node
type Landmark struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Location  Point   `json:"location"`
	MinRTT    float64 `json:"min_rtt_ms"`
	MaxRadius float64 `json:"max_radius_km"`
	Samples   int     `json:"samples"`
}

// MultilaterationResult contains the computed target area and coordinates
type MultilaterationResult struct {
	EstimatedPoint Point      `json:"estimated_point"`
	ConfidenceKm   float64    `json:"confidence_radius_km"`
	UsedLandmarks  []Landmark `json:"used_landmarks"`
	PolygonBounds  []Point    `json:"polygon_bounds,omitempty"`
}

// SolveCentroidLeastSquares estimates the geographic coordinates using weighted multilateration
func SolveCentroidLeastSquares(landmarks []Landmark) MultilaterationResult {
	if len(landmarks) == 0 {
		return MultilaterationResult{}
	}

	if len(landmarks) == 1 {
		return MultilaterationResult{
			EstimatedPoint: landmarks[0].Location,
			ConfidenceKm:   landmarks[0].MaxRadius,
			UsedLandmarks:  landmarks,
		}
	}

	// Grid search optimization with iterative refinement (simulated annealing / gradient descent over geoid)
	// 1. Initial bounding box from landmark locations
	minLat, maxLat := 90.0, -90.0
	minLon, maxLon := 180.0, -180.0
	for _, l := range landmarks {
		if l.Location.Lat < minLat {
			minLat = l.Location.Lat
		}
		if l.Location.Lat > maxLat {
			maxLat = l.Location.Lat
		}
		if l.Location.Lon < minLon {
			minLon = l.Location.Lon
		}
		if l.Location.Lon > maxLon {
			maxLon = l.Location.Lon
		}
	}

	// Expand bounding box slightly based on radius
	maxR := 0.0
	for _, l := range landmarks {
		if l.MaxRadius > maxR {
			maxR = l.MaxRadius
		}
	}
	degPad := (maxR / EarthRadiusKm) * (180.0 / math.Pi)
	minLat = math.Max(-85.0, minLat-degPad)
	maxLat = math.Min(85.0, maxLat+degPad)
	minLon = math.Max(-180.0, minLon-degPad)
	maxLon = math.Min(180.0, maxLon+degPad)

	bestPoint := Point{Lat: (minLat + maxLat) / 2.0, Lon: (minLon + maxLon) / 2.0}
	minLoss := math.MaxFloat64

	// Multi-resolution search
	steps := 40
	latStep := (maxLat - minLat) / float64(steps)
	lonStep := (maxLon - minLon) / float64(steps)

	for i := 0; i <= steps; i++ {
		curLat := minLat + float64(i)*latStep
		for j := 0; j <= steps; j++ {
			curLon := minLon + float64(j)*lonStep
			candidate := Point{Lat: curLat, Lon: curLon}

			loss := computeCBGLoss(candidate, landmarks)
			if loss < minLoss {
				minLoss = loss
				bestPoint = candidate
			}
		}
	}

	// Refinement pass around bestPoint
	refineLatStep := latStep / 10.0
	refineLonStep := lonStep / 10.0
	centerLat := bestPoint.Lat
	centerLon := bestPoint.Lon

	for i := -15; i <= 15; i++ {
		curLat := centerLat + float64(i)*refineLatStep
		if curLat < -90 || curLat > 90 {
			continue
		}
		for j := -15; j <= 15; j++ {
			curLon := centerLon + float64(j)*refineLonStep
			if curLon < -180 || curLon > 180 {
				continue
			}
			candidate := Point{Lat: curLat, Lon: curLon}
			loss := computeCBGLoss(candidate, landmarks)
			if loss < minLoss {
				minLoss = loss
				bestPoint = candidate
			}
		}
	}

	// Calculate average distance error / confidence radius
	confidence := 0.0
	for _, l := range landmarks {
		dist := DistanceHaversine(bestPoint, l.Location)
		diff := math.Abs(dist - (l.MaxRadius * 0.7)) // expected distance within constraint
		confidence += diff
	}
	confidence = confidence / float64(len(landmarks))

	return MultilaterationResult{
		EstimatedPoint: bestPoint,
		ConfidenceKm:   math.Max(15.0, confidence),
		UsedLandmarks:  landmarks,
	}
}

// computeCBGLoss penalizes candidates that violate the physical speed-of-light constraint (distance > maxRadius)
// and rewards points that best fit all distance arcs.
func computeCBGLoss(p Point, landmarks []Landmark) float64 {
	totalLoss := 0.0
	for _, l := range landmarks {
		dist := DistanceHaversine(p, l.Location)
		if dist > l.MaxRadius {
			// Hard physical violation penalty (light speed in fiber violation)
			violation := dist - l.MaxRadius
			totalLoss += (violation * violation) * 1000.0
		} else {
			// Soft fit penalty
			diff := dist - (l.MaxRadius * 0.65)
			totalLoss += diff * diff
		}
	}
	return totalLoss
}
