package proximity

import (
	"context"
	"errors"

	"github.com/nagicantsleep/k-map/internal/api"
)

// DefaultThresholdMeters is the default proximity threshold when the caller
// does not specify one.
const DefaultThresholdMeters = 100.0

// ErrNoTargetMatch is returned when the target query yields no geocoding results.
var ErrNoTargetMatch = errors.New("no geocoding results for target query")

// Service implements proximity validation by geocoding a target query,
// computing geodesic distance, and applying a threshold rule.
type Service struct {
	geocoder api.Geocoder
}

// NewService creates a new proximity service.
func NewService(geocoder api.Geocoder) *Service {
	return &Service{geocoder: geocoder}
}

// Result holds the outcome of a proximity check.
type Result struct {
	IsNear          bool
	DistanceMeters  float64
	ThresholdMeters float64
	TargetMatch     *api.GeocodeResult
}

// Check geocodes the target query, computes the geodesic distance from
// (lat, lng) to the best candidate, and returns whether the point is near.
func (s *Service) Check(ctx context.Context, lat, lng float64, targetQuery string, thresholdMeters float64) (*Result, error) {
	if thresholdMeters <= 0 {
		thresholdMeters = DefaultThresholdMeters
	}

	results, err := s.geocoder.Search(ctx, targetQuery, 1)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrNoTargetMatch
	}

	best := results[0]
	distance := Haversine(lat, lng, best.Latitude, best.Longitude)

	return &Result{
		IsNear:          distance <= thresholdMeters,
		DistanceMeters:  distance,
		ThresholdMeters: thresholdMeters,
		TargetMatch:     &best,
	}, nil
}
