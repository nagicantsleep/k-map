//go:build integration

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nagicantsleep/k-map/internal/api"
	"github.com/nagicantsleep/k-map/internal/geocode"
)

// newTestServer creates an HTTP test server backed by a real Nominatim instance.
// Set KMAP_NOMINATIM_URL env var to point to your local Nominatim.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	nominatimURL := "http://localhost:8081"
	client := geocode.NewNominatimClient(nominatimURL, 10*time.Second)

	handler := api.NewHandler(api.HandlerOptions{
		Geocoder: client,
	})

	return httptest.NewServer(handler)
}

// -- Forward Geocode Tests --

func TestIntegration_ForwardGeocode_ValidQuery(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	body := `{"query":"Mairie de Monaco","limit":3}`
	resp, err := http.Post(srv.URL+"/v1/geocode/forward", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result api.ForwardGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if result.Query != "Mairie de Monaco" {
		t.Errorf("expected query echoed back, got %q", result.Query)
	}

	if len(result.Results) == 0 {
		t.Fatal("expected at least one result")
	}

	first := result.Results[0]
	if first.FormattedAddress == "" {
		t.Error("expected non-empty formatted address")
	}

	if first.Source != "osm" {
		t.Errorf("expected source 'osm', got %q", first.Source)
	}

	// Monaco coordinates should be roughly 43.7N, 7.4E
	if first.Latitude < 43.0 || first.Latitude > 44.0 {
		t.Errorf("latitude out of expected range: %f", first.Latitude)
	}

	if first.Longitude < 7.0 || first.Longitude > 8.0 {
		t.Errorf("longitude out of expected range: %f", first.Longitude)
	}
}

func TestIntegration_ForwardGeocode_NoMatch(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	body := `{"query":"zzz_completely_nonexistent_address_xyz_12345"}`
	resp, err := http.Post(srv.URL+"/v1/geocode/forward", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (empty results), got %d", resp.StatusCode)
	}

	var result api.ForwardGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(result.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(result.Results))
	}
}

func TestIntegration_ForwardGeocode_InvalidInput(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Missing query field
	body := `{"limit":5}`
	resp, err := http.Post(srv.URL+"/v1/geocode/forward", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// -- Reverse Geocode Tests --

func TestIntegration_ReverseGeocode_ValidCoordinates(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Monaco Palace coordinates
	body := `{"latitude":43.7311,"longitude":7.4197}`
	resp, err := http.Post(srv.URL+"/v1/geocode/reverse", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result api.ReverseGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if result.Latitude != 43.7311 {
		t.Errorf("expected latitude 43.7311, got %f", result.Latitude)
	}

	if result.Result == nil {
		t.Fatal("expected non-nil result for Monaco coordinates")
	}

	if result.Result.FormattedAddress == "" {
		t.Error("expected non-empty formatted address")
	}

	fmt.Printf("Reverse geocode result: %s\n", result.Result.FormattedAddress)
}

func TestIntegration_ReverseGeocode_NoMatch(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Middle of the ocean
	body := `{"latitude":0.0001,"longitude":0.0001}`
	resp, err := http.Post(srv.URL+"/v1/geocode/reverse", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should still return 200 with null result (or a result if Nominatim finds something nearby)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_ReverseGeocode_InvalidCoordinates(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	body := `{"latitude":91.0,"longitude":0.0}`
	resp, err := http.Post(srv.URL+"/v1/geocode/reverse", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
