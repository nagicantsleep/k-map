package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockGeocoder implements the Geocoder interface for testing.
type mockGeocoder struct {
	searchResults []GeocodeResult
	searchErr     error
	reverseResult *GeocodeResult
	reverseErr    error
}

func (m *mockGeocoder) Search(_ context.Context, _ string, _ int) ([]GeocodeResult, error) {
	return m.searchResults, m.searchErr
}

func (m *mockGeocoder) Reverse(_ context.Context, _, _ float64) (*GeocodeResult, error) {
	return m.reverseResult, m.reverseErr
}

func TestForwardGeocodeHandler_Success(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{
		searchResults: []GeocodeResult{
			{
				FormattedAddress: "Test Address",
				Latitude:         37.422,
				Longitude:        -122.084,
				Confidence:       0.9,
				Source:           "osm",
			},
		},
	}

	handler := forwardGeocodeHandler(geo)
	body := `{"query":"test address","limit":5}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/forward", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ForwardGeocodeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Query != "test address" {
		t.Errorf("expected query 'test address', got %q", resp.Query)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].FormattedAddress != "Test Address" {
		t.Errorf("unexpected formatted address: %s", resp.Results[0].FormattedAddress)
	}
}

func TestForwardGeocodeHandler_EmptyResults(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{searchResults: nil}

	handler := forwardGeocodeHandler(geo)
	body := `{"query":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/forward", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ForwardGeocodeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Results == nil {
		t.Fatal("expected non-nil results array")
	}

	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

func TestForwardGeocodeHandler_MissingQuery(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{}

	handler := forwardGeocodeHandler(geo)
	body := `{"limit":5}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/forward", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestForwardGeocodeHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{}

	handler := forwardGeocodeHandler(geo)
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/forward", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestForwardGeocodeHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{}

	handler := forwardGeocodeHandler(geo)
	req := httptest.NewRequest(http.MethodGet, "/v1/geocode/forward", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestForwardGeocodeHandler_GeocoderError(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{searchErr: errors.New("nominatim down")}

	handler := forwardGeocodeHandler(geo)
	body := `{"query":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/forward", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestForwardGeocodeHandler_DefaultLimit(t *testing.T) {
	t.Parallel()

	var capturedLimit int
	geo := &mockGeocoder{}

	// Use a custom mock to capture the limit
	handler := forwardGeocodeHandler(&limitCapturingGeocoder{
		limit:   &capturedLimit,
		results: []GeocodeResult{},
	})
	body := `{"query":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/forward", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	_ = geo // unused but kept for clarity
	handler.ServeHTTP(rec, req)

	if capturedLimit != 5 {
		t.Errorf("expected default limit 5, got %d", capturedLimit)
	}
}

type limitCapturingGeocoder struct {
	limit   *int
	results []GeocodeResult
}

func (g *limitCapturingGeocoder) Search(_ context.Context, _ string, limit int) ([]GeocodeResult, error) {
	*g.limit = limit
	return g.results, nil
}

func (g *limitCapturingGeocoder) Reverse(_ context.Context, _, _ float64) (*GeocodeResult, error) {
	return nil, nil
}

func TestReverseGeocodeHandler_Success(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{
		reverseResult: &GeocodeResult{
			FormattedAddress: "1600 Amphitheatre Parkway, Mountain View, CA",
			Latitude:         37.422,
			Longitude:        -122.084,
			Confidence:       0.9,
			Source:           "osm",
			PlaceType:        "house",
		},
	}

	handler := reverseGeocodeHandler(geo)
	body := `{"latitude":37.422,"longitude":-122.084}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/reverse", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ReverseGeocodeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Latitude != 37.422 {
		t.Errorf("expected latitude 37.422, got %f", resp.Latitude)
	}

	if resp.Result == nil {
		t.Fatal("expected non-nil result")
	}

	if resp.Result.FormattedAddress != "1600 Amphitheatre Parkway, Mountain View, CA" {
		t.Errorf("unexpected formatted address: %s", resp.Result.FormattedAddress)
	}
}

func TestReverseGeocodeHandler_NoMatch(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{reverseResult: nil}

	handler := reverseGeocodeHandler(geo)
	body := `{"latitude":0.0,"longitude":0.0}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/reverse", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ReverseGeocodeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Result != nil {
		t.Errorf("expected nil result, got %+v", resp.Result)
	}
}

func TestReverseGeocodeHandler_InvalidCoordinates(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{}

	handler := reverseGeocodeHandler(geo)
	body := `{"latitude":91.0,"longitude":0.0}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/reverse", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestReverseGeocodeHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{}

	handler := reverseGeocodeHandler(geo)
	req := httptest.NewRequest(http.MethodGet, "/v1/geocode/reverse", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestReverseGeocodeHandler_GeocoderError(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{reverseErr: errors.New("nominatim down")}

	handler := reverseGeocodeHandler(geo)
	body := `{"latitude":37.422,"longitude":-122.084}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/reverse", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestReverseGeocodeHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	geo := &mockGeocoder{}

	handler := reverseGeocodeHandler(geo)
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/reverse", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// mockProximityChecker implements ProximityChecker for testing.
type mockProximityChecker struct {
	result *ProximityResult
	err    error
}

func (m *mockProximityChecker) Check(_ context.Context, _, _ float64, _ string, _ float64) (*ProximityResult, error) {
	return m.result, m.err
}

func TestProximityHandler_NearSuccess(t *testing.T) {
	t.Parallel()

	checker := &mockProximityChecker{
		result: &ProximityResult{
			IsNear:          true,
			DistanceMeters:  42.5,
			ThresholdMeters: 100,
			TargetMatch: &GeocodeResult{
				FormattedAddress: "1600 Amphitheatre Parkway",
				Latitude:         37.422,
				Longitude:        -122.084,
				Confidence:       0.9,
				Source:           "osm",
			},
		},
	}

	handler := proximityHandler(checker)
	body := `{"latitude":37.42195,"longitude":-122.08405,"target_query":"1600 Amphitheatre Parkway","threshold_meters":100}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/proximity", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ProximityResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.IsNear {
		t.Error("expected is_near=true")
	}
	if resp.DistanceMeters != 42.5 {
		t.Errorf("expected distance 42.5, got %f", resp.DistanceMeters)
	}
	if resp.TargetMatch == nil {
		t.Fatal("expected non-nil target_match")
	}
}

func TestProximityHandler_NotNear(t *testing.T) {
	t.Parallel()

	checker := &mockProximityChecker{
		result: &ProximityResult{
			IsNear:          false,
			DistanceMeters:  5000,
			ThresholdMeters: 100,
			TargetMatch: &GeocodeResult{
				FormattedAddress: "Far Away Place",
				Latitude:         38.0,
				Longitude:        -123.0,
				Confidence:       0.7,
				Source:           "osm",
			},
		},
	}

	handler := proximityHandler(checker)
	body := `{"latitude":37.42195,"longitude":-122.08405,"target_query":"Far Away Place","threshold_meters":100}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/proximity", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ProximityResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.IsNear {
		t.Error("expected is_near=false")
	}
}

func TestProximityHandler_InvalidCoordinate(t *testing.T) {
	t.Parallel()

	checker := &mockProximityChecker{}
	handler := proximityHandler(checker)
	body := `{"latitude":91.0,"longitude":0.0,"target_query":"test","threshold_meters":100}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/proximity", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestProximityHandler_MissingTargetQuery(t *testing.T) {
	t.Parallel()

	checker := &mockProximityChecker{}
	handler := proximityHandler(checker)
	body := `{"latitude":37.422,"longitude":-122.084,"threshold_meters":100}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/proximity", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestProximityHandler_CheckerError(t *testing.T) {
	t.Parallel()

	checker := &mockProximityChecker{err: errors.New("geocoder down")}
	handler := proximityHandler(checker)
	body := `{"latitude":37.422,"longitude":-122.084,"target_query":"test place","threshold_meters":100}`
	req := httptest.NewRequest(http.MethodPost, "/v1/geocode/proximity", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestProximityHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	checker := &mockProximityChecker{}
	handler := proximityHandler(checker)
	req := httptest.NewRequest(http.MethodGet, "/v1/geocode/proximity", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
