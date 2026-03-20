//go:build integration

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/nagicantsleep/k-map/internal/api"
	"github.com/nagicantsleep/k-map/internal/geocode"
)

// goldenFixtures mirrors the structure of test/fixtures/monaco_golden.json.
type goldenFixtures struct {
	Description string              `json:"description"`
	Forward     []forwardFixture    `json:"forward"`
	Reverse     []reverseFixture    `json:"reverse"`
	Proximity   []proximityFixture  `json:"proximity"`
}

type forwardFixture struct {
	ID              string  `json:"id"`
	Query           string  `json:"query"`
	ExpectLatMin    float64 `json:"expect_lat_min"`
	ExpectLatMax    float64 `json:"expect_lat_max"`
	ExpectLonMin    float64 `json:"expect_lon_min"`
	ExpectLonMax    float64 `json:"expect_lon_max"`
	ExpectMinResult int     `json:"expect_min_results"`
}

type reverseFixture struct {
	ID            string  `json:"id"`
	Description   string  `json:"description"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	ExpectResult  bool    `json:"expect_result"`
}

type proximityFixture struct {
	ID              string  `json:"id"`
	Description     string  `json:"description"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TargetQuery     string  `json:"target_query"`
	ThresholdMeters float64 `json:"threshold_meters"`
	ExpectIsNear    bool    `json:"expect_is_near"`
}

func loadGoldenFixtures(t *testing.T) goldenFixtures {
	t.Helper()

	path := "fixtures/monaco_golden.json"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden fixture file %s: %v", path, err)
	}

	var fixtures goldenFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("failed to parse golden fixture file: %v", err)
	}

	return fixtures
}

func newGoldenTestServer(t *testing.T) string {
	t.Helper()

	// Allow override via env var; default to the local stack.
	baseURL := os.Getenv("KMAP_API_URL")
	if baseURL != "" {
		return baseURL
	}

	nominatimURL := os.Getenv("KMAP_NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "http://localhost:8081"
	}

	client := geocode.NewNominatimClient(nominatimURL, 10*time.Second)
	handler := api.NewHandler(api.HandlerOptions{
		Geocoder: client,
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestGoldenDataset_Forward validates all forward geocoding fixture cases.
func TestGoldenDataset_Forward(t *testing.T) {
	baseURL := newGoldenTestServer(t)
	fixtures := loadGoldenFixtures(t)

	passed, failed := 0, 0
	for _, tc := range fixtures.Forward {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			body, _ := json.Marshal(api.ForwardGeocodeRequest{Query: tc.Query, Limit: 5})
			resp, err := http.Post(baseURL+"/v1/geocode/forward", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("[%s] request failed: %v", tc.ID, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("[%s] expected 200, got %d", tc.ID, resp.StatusCode)
				failed++
				return
			}

			var result api.ForwardGeocodeResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("[%s] decode failed: %v", tc.ID, err)
			}

			if len(result.Results) < tc.ExpectMinResult {
				t.Errorf("[%s] expected >= %d results, got %d", tc.ID, tc.ExpectMinResult, len(result.Results))
				failed++
				return
			}

			first := result.Results[0]
			if first.Latitude < tc.ExpectLatMin || first.Latitude > tc.ExpectLatMax {
				t.Errorf("[%s] latitude %f out of expected range [%f, %f]", tc.ID, first.Latitude, tc.ExpectLatMin, tc.ExpectLatMax)
				failed++
				return
			}

			if first.Longitude < tc.ExpectLonMin || first.Longitude > tc.ExpectLonMax {
				t.Errorf("[%s] longitude %f out of expected range [%f, %f]", tc.ID, first.Longitude, tc.ExpectLonMin, tc.ExpectLonMax)
				failed++
				return
			}

			fmt.Printf("PASS [%s] query=%q lat=%.4f lon=%.4f addr=%q\n",
				tc.ID, tc.Query, first.Latitude, first.Longitude, first.FormattedAddress)
			passed++
		})
	}

	t.Logf("Forward geocoding: %d passed, %d failed out of %d cases", passed, failed, len(fixtures.Forward))
}

// TestGoldenDataset_Reverse validates all reverse geocoding fixture cases.
func TestGoldenDataset_Reverse(t *testing.T) {
	baseURL := newGoldenTestServer(t)
	fixtures := loadGoldenFixtures(t)

	passed, failed := 0, 0
	for _, tc := range fixtures.Reverse {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			body, _ := json.Marshal(api.ReverseGeocodeRequest{Latitude: tc.Latitude, Longitude: tc.Longitude})
			resp, err := http.Post(baseURL+"/v1/geocode/reverse", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("[%s] request failed: %v", tc.ID, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("[%s] expected 200, got %d", tc.ID, resp.StatusCode)
				failed++
				return
			}

			var result api.ReverseGeocodeResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("[%s] decode failed: %v", tc.ID, err)
			}

			if tc.ExpectResult && result.Result == nil {
				t.Errorf("[%s] expected a result for (%.4f, %.4f) but got nil", tc.ID, tc.Latitude, tc.Longitude)
				failed++
				return
			}

			if tc.ExpectResult {
				fmt.Printf("PASS [%s] %s -> %q\n", tc.ID, tc.Description, result.Result.FormattedAddress)
			} else {
				fmt.Printf("PASS [%s] %s -> no result (expected)\n", tc.ID, tc.Description)
			}
			passed++
		})
	}

	t.Logf("Reverse geocoding: %d passed, %d failed out of %d cases", passed, failed, len(fixtures.Reverse))
}

// TestGoldenDataset_Proximity validates all proximity fixture cases.
func TestGoldenDataset_Proximity(t *testing.T) {
	baseURL := newGoldenTestServer(t)
	fixtures := loadGoldenFixtures(t)

	passed, failed := 0, 0
	for _, tc := range fixtures.Proximity {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			body, _ := json.Marshal(api.ProximityRequest{
				Latitude:        tc.Latitude,
				Longitude:       tc.Longitude,
				TargetQuery:     tc.TargetQuery,
				ThresholdMeters: tc.ThresholdMeters,
			})
			resp, err := http.Post(baseURL+"/v1/geocode/proximity", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("[%s] request failed: %v", tc.ID, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("[%s] expected 200, got %d", tc.ID, resp.StatusCode)
				failed++
				return
			}

			var result api.ProximityResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("[%s] decode failed: %v", tc.ID, err)
			}

			if result.IsNear != tc.ExpectIsNear {
				t.Errorf("[%s] %s: expected is_near=%v, got is_near=%v (distance=%.1fm, threshold=%.1fm)",
					tc.ID, tc.Description, tc.ExpectIsNear, result.IsNear, result.DistanceMeters, result.ThresholdMeters)
				failed++
				return
			}

			fmt.Printf("PASS [%s] %s -> is_near=%v distance=%.1fm threshold=%.1fm\n",
				tc.ID, tc.Description, result.IsNear, result.DistanceMeters, result.ThresholdMeters)
			passed++
		})
	}

	t.Logf("Proximity validation: %d passed, %d failed out of %d cases", passed, failed, len(fixtures.Proximity))
}
