package geocode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNominatimClient_Search(t *testing.T) {
	t.Parallel()

	mockResults := []nominatimResult{
		{
			PlaceID:     123,
			Lat:         "37.422",
			Lon:         "-122.084",
			DisplayName: "1600 Amphitheatre Parkway, Mountain View, CA, USA",
			Type:        "house",
			Importance:  0.9,
			Address: &nominatimAddress{
				HouseNumber: "1600",
				Road:        "Amphitheatre Parkway",
				City:        "Mountain View",
				State:       "CA",
				Postcode:    "94043",
				Country:     "United States",
				CountryCode: "US",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("expected path /search, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("q") != "1600 Amphitheatre Parkway" {
			t.Errorf("expected query '1600 Amphitheatre Parkway', got %s", r.URL.Query().Get("q"))
		}

		if r.URL.Query().Get("format") != "json" {
			t.Errorf("expected format json, got %s", r.URL.Query().Get("format"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResults)
	}))
	defer server.Close()

	client := NewNominatimClient(server.URL, 5*time.Second)
	results, err := client.Search(context.Background(), "1600 Amphitheatre Parkway", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].FormattedAddress != "1600 Amphitheatre Parkway, Mountain View, CA, USA" {
		t.Errorf("unexpected formatted address: %s", results[0].FormattedAddress)
	}

	if results[0].Latitude != 37.422 {
		t.Errorf("expected latitude 37.422, got %f", results[0].Latitude)
	}

	if results[0].Components.City != "Mountain View" {
		t.Errorf("expected city 'Mountain View', got %s", results[0].Components.City)
	}
}

func TestNominatimClient_Search_EmptyResults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]nominatimResult{})
	}))
	defer server.Close()

	client := NewNominatimClient(server.URL, 5*time.Second)
	results, err := client.Search(context.Background(), "nonexistent place", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results != nil {
		t.Errorf("expected nil results for empty response, got %d results", len(results))
	}
}

func TestNominatimClient_Reverse(t *testing.T) {
	t.Parallel()

	mockResult := nominatimResult{
		PlaceID:     456,
		Lat:         "37.422",
		Lon:         "-122.084",
		DisplayName: "1600 Amphitheatre Parkway, Mountain View, CA, USA",
		Type:        "house",
		Importance:  0.9,
		Address: &nominatimAddress{
			HouseNumber: "1600",
			Road:        "Amphitheatre Parkway",
			City:        "Mountain View",
			State:       "CA",
			Country:     "United States",
			CountryCode: "US",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reverse" {
			t.Errorf("expected path /reverse, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("lat") != "37.422" {
			t.Errorf("expected lat 37.422, got %s", r.URL.Query().Get("lat"))
		}

		if r.URL.Query().Get("lon") != "-122.084" {
			t.Errorf("expected lon -122.084, got %s", r.URL.Query().Get("lon"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResult)
	}))
	defer server.Close()

	client := NewNominatimClient(server.URL, 5*time.Second)
	result, err := client.Reverse(context.Background(), 37.422, -122.084)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Latitude != 37.422 {
		t.Errorf("expected latitude 37.422, got %f", result.Latitude)
	}

	if result.Components.StreetNumber != "1600" {
		t.Errorf("expected street_number '1600', got %s", result.Components.StreetNumber)
	}
}

func TestNominatimClient_Reverse_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(nominatimResult{})
	}))
	defer server.Close()

	client := NewNominatimClient(server.URL, 5*time.Second)
	result, err := client.Reverse(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for not found, got %+v", result)
	}
}

func TestNominatimClient_Search_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := NewNominatimClient(server.URL, 10*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := client.Search(ctx, "test", 10)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestNominatimClient_Search_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewNominatimClient(server.URL, 5*time.Second)
	_, err := client.Search(context.Background(), "test", 10)
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestNominatimClient_NormalizeResult_TownFallback(t *testing.T) {
	t.Parallel()

	result := nominatimResult{
		PlaceID:     789,
		Lat:         "48.8566",
		Lon:         "2.3522",
		DisplayName: "Paris, France",
		Type:        "city",
		Importance:  0.8,
		Address: &nominatimAddress{
			Town:        "Paris",
			Country:     "France",
			CountryCode: "FR",
		},
	}

	client := &NominatimClient{}
	normalized := client.normalizeResult(result)

	if normalized.Components.City != "Paris" {
		t.Errorf("expected city 'Paris' from town, got %s", normalized.Components.City)
	}
}

func TestNominatimClient_NormalizeResult_VillageFallback(t *testing.T) {
	t.Parallel()

	result := nominatimResult{
		PlaceID:     789,
		Lat:         "45.0",
		Lon:         "5.0",
		DisplayName: "Small Village, France",
		Type:        "village",
		Importance:  0.5,
		Address: &nominatimAddress{
			Village:     "Small Village",
			Country:     "France",
			CountryCode: "FR",
		},
	}

	client := &NominatimClient{}
	normalized := client.normalizeResult(result)

	if normalized.Components.City != "Small Village" {
		t.Errorf("expected city 'Small Village' from village, got %s", normalized.Components.City)
	}
}
