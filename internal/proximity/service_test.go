package proximity

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/nagicantsleep/k-map/internal/api"
)

// mockGeocoder implements api.Geocoder for proximity service tests.
type mockGeocoder struct {
	results []api.GeocodeResult
	err     error
}

func (m *mockGeocoder) Search(_ context.Context, _ string, _ int) ([]api.GeocodeResult, error) {
	return m.results, m.err
}

func (m *mockGeocoder) Reverse(_ context.Context, _, _ float64) (*api.GeocodeResult, error) {
	return nil, nil
}

// Labeled near/not-near fixtures.
var proximityFixtures = []struct {
	name            string
	lat, lng        float64
	targetLat       float64
	targetLng       float64
	thresholdMeters float64
	wantNear        bool
	description     string
}{
	{
		name:            "near: same location",
		lat: 37.42200, lng: -122.08400,
		targetLat: 37.42200, targetLng: -122.08400,
		thresholdMeters: 100,
		wantNear:        true,
		description:     "Exact same point should always be near",
	},
	{
		name:            "near: ~50m apart within 100m threshold",
		lat: 37.42200, lng: -122.08400,
		targetLat: 37.42245, targetLng: -122.08400,
		thresholdMeters: 100,
		wantNear:        true,
		description:     "Point ~50m north should be within 100m threshold",
	},
	{
		name:            "near: just within threshold boundary",
		lat: 37.42200, lng: -122.08400,
		targetLat: 37.42289, targetLng: -122.08400,
		thresholdMeters: 100,
		wantNear:        true,
		description:     "Point ~99m north should be within 100m threshold",
	},
	{
		name:            "not-near: ~500m apart with 100m threshold",
		lat: 37.42200, lng: -122.08400,
		targetLat: 37.42650, targetLng: -122.08400,
		thresholdMeters: 100,
		wantNear:        false,
		description:     "Point ~500m north should exceed 100m threshold",
	},
	{
		name:            "not-near: different cities",
		lat: 37.7749, lng: -122.4194,
		targetLat: 34.0522, targetLng: -118.2437,
		thresholdMeters: 1000,
		wantNear:        false,
		description:     "SF to LA should exceed any reasonable threshold",
	},
	{
		name:            "near: large threshold encompasses distant point",
		lat: 37.42200, lng: -122.08400,
		targetLat: 37.42650, targetLng: -122.08400,
		thresholdMeters: 1000,
		wantNear:        true,
		description:     "~500m apart should be within 1000m threshold",
	},
}

func TestService_Check_NearNotNear(t *testing.T) {
	for _, tt := range proximityFixtures {
		t.Run(tt.name, func(t *testing.T) {
			geo := &mockGeocoder{
				results: []api.GeocodeResult{
					{
						FormattedAddress: "Test Target",
						Latitude:         tt.targetLat,
						Longitude:        tt.targetLng,
						Confidence:       0.9,
						Source:           "osm",
					},
				},
			}

			svc := NewService(geo)
			result, err := svc.Check(context.Background(), tt.lat, tt.lng, "test target", tt.thresholdMeters)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.IsNear != tt.wantNear {
				t.Errorf("%s: got is_near=%v, want %v (distance=%.2fm, threshold=%.2fm)",
					tt.description, result.IsNear, tt.wantNear, result.DistanceMeters, result.ThresholdMeters)
			}

			if result.ThresholdMeters != tt.thresholdMeters {
				t.Errorf("expected threshold %f, got %f", tt.thresholdMeters, result.ThresholdMeters)
			}

			if result.TargetMatch == nil {
				t.Fatal("expected non-nil target match")
			}
		})
	}
}

func TestService_Check_DefaultThreshold(t *testing.T) {
	geo := &mockGeocoder{
		results: []api.GeocodeResult{
			{
				FormattedAddress: "Nearby",
				Latitude:         37.42200,
				Longitude:        -122.08400,
				Confidence:       0.9,
				Source:           "osm",
			},
		},
	}

	svc := NewService(geo)
	result, err := svc.Check(context.Background(), 37.42200, -122.08400, "nearby", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ThresholdMeters != DefaultThresholdMeters {
		t.Errorf("expected default threshold %f, got %f", DefaultThresholdMeters, result.ThresholdMeters)
	}
}

func TestService_Check_NoResults(t *testing.T) {
	geo := &mockGeocoder{results: nil}

	svc := NewService(geo)
	_, err := svc.Check(context.Background(), 37.42200, -122.08400, "nonexistent place", 100)

	if !errors.Is(err, ErrNoTargetMatch) {
		t.Errorf("expected ErrNoTargetMatch, got %v", err)
	}
}

func TestService_Check_GeocoderError(t *testing.T) {
	geo := &mockGeocoder{err: errors.New("nominatim down")}

	svc := NewService(geo)
	_, err := svc.Check(context.Background(), 37.42200, -122.08400, "test", 100)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Check_Deterministic(t *testing.T) {
	geo := &mockGeocoder{
		results: []api.GeocodeResult{
			{
				FormattedAddress: "Stable Target",
				Latitude:         37.42245,
				Longitude:        -122.08400,
				Confidence:       0.9,
				Source:           "osm",
			},
		},
	}

	svc := NewService(geo)
	ctx := context.Background()

	result1, err := svc.Check(ctx, 37.42200, -122.08400, "stable target", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result2, err := svc.Check(ctx, 37.42200, -122.08400, "stable target", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result1.IsNear != result2.IsNear {
		t.Error("is_near not deterministic")
	}
	if math.Abs(result1.DistanceMeters-result2.DistanceMeters) > 0.001 {
		t.Errorf("distance not deterministic: %f vs %f", result1.DistanceMeters, result2.DistanceMeters)
	}
	if result1.ThresholdMeters != result2.ThresholdMeters {
		t.Error("threshold not deterministic")
	}
}
