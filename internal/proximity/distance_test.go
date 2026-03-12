package proximity

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name           string
		lat1, lng1     float64
		lat2, lng2     float64
		wantMeters     float64
		toleranceMeters float64
	}{
		{
			name:            "same point",
			lat1: 37.7749, lng1: -122.4194,
			lat2: 37.7749, lng2: -122.4194,
			wantMeters:     0,
			toleranceMeters: 0.01,
		},
		{
			name:            "New York to Los Angeles (~3944 km)",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 34.0522, lng2: -118.2437,
			wantMeters:     3_944_000,
			toleranceMeters: 10_000, // 10 km tolerance for long distance
		},
		{
			name:            "London to Paris (~344 km)",
			lat1: 51.5074, lng1: -0.1278,
			lat2: 48.8566, lng2: 2.3522,
			wantMeters:     343_500,
			toleranceMeters: 1_000,
		},
		{
			name:            "short distance (~50 meters)",
			lat1: 37.42200, lng1: -122.08400,
			lat2: 37.42245, lng2: -122.08400,
			wantMeters:     50,
			toleranceMeters: 1,
		},
		{
			name:            "equator crossing",
			lat1: 0.001, lng1: 0.0,
			lat2: -0.001, lng2: 0.0,
			wantMeters:     222,
			toleranceMeters: 5,
		},
		{
			name:            "antipodal points",
			lat1: 0, lng1: 0,
			lat2: 0, lng2: 180,
			wantMeters:     math.Pi * EarthRadiusMeters,
			toleranceMeters: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Haversine(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			diff := math.Abs(got - tt.wantMeters)
			if diff > tt.toleranceMeters {
				t.Errorf("Haversine(%f, %f, %f, %f) = %f, want ~%f (diff %f > tolerance %f)",
					tt.lat1, tt.lng1, tt.lat2, tt.lng2, got, tt.wantMeters, diff, tt.toleranceMeters)
			}
		})
	}
}

func TestHaversineSymmetry(t *testing.T) {
	d1 := Haversine(40.7128, -74.0060, 34.0522, -118.2437)
	d2 := Haversine(34.0522, -118.2437, 40.7128, -74.0060)

	if d1 != d2 {
		t.Errorf("Haversine not symmetric: %f != %f", d1, d2)
	}
}
