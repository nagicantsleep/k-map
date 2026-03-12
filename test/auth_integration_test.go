package auth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nagicantsleep/k-map/internal/api"
	"github.com/nagicantsleep/k-map/internal/auth"
)

// --- Test doubles ---

type stubGeocoder struct{}

func (s *stubGeocoder) Search(_ context.Context, _ string, _ int) ([]api.GeocodeResult, error) {
	return []api.GeocodeResult{{FormattedAddress: "stub"}}, nil
}

func (s *stubGeocoder) Reverse(_ context.Context, _, _ float64) (*api.GeocodeResult, error) {
	return &api.GeocodeResult{FormattedAddress: "stub"}, nil
}

type stubProximity struct{}

func (s *stubProximity) Check(_ context.Context, _, _ float64, _ string, threshold float64) (*api.ProximityResult, error) {
	return &api.ProximityResult{
		IsNear:          true,
		DistanceMeters:  10,
		ThresholdMeters: threshold,
	}, nil
}

type inMemoryKeyLookup struct {
	keys map[string]*auth.APIKey
}

func (m *inMemoryKeyLookup) LookupByHash(_ context.Context, keyHash string) (*auth.APIKey, error) {
	k, ok := m.keys[keyHash]
	if !ok {
		return nil, sql.ErrNoRows
	}

	return k, nil
}

func (m *inMemoryKeyLookup) TouchAPIKey(_ context.Context, _ string) error {
	return nil
}

type inMemoryRateLimiter struct {
	allowed bool
}

func (m *inMemoryRateLimiter) Allow(_ context.Context, _ string) (bool, error) {
	return m.allowed, nil
}

type inMemoryUsageRecorder struct {
	records []usageRecord
}

type usageRecord struct {
	TenantID   string
	Endpoint   string
	RequestID  string
	StatusCode int
	LatencyMs  int
}

func (m *inMemoryUsageRecorder) Record(_ context.Context, tenantID, endpoint, requestID string, statusCode, latencyMs int) error {
	m.records = append(m.records, usageRecord{
		TenantID:   tenantID,
		Endpoint:   endpoint,
		RequestID:  requestID,
		StatusCode: statusCode,
		LatencyMs:  latencyMs,
	})

	return nil
}

// --- Test helpers ---

func buildTestServer(lookup auth.KeyLookup, limiter auth.RateLimitChecker, recorder auth.UsageRecorderInterface) *httptest.Server {
	handler := api.NewHandler(api.HandlerOptions{
		Geocoder:            &stubGeocoder{},
		Proximity:           &stubProximity{},
		AuthMiddleware:      auth.AuthMiddleware(lookup),
		RateLimitMiddleware: auth.RateLimitMiddleware(limiter),
		UsageMiddleware:     auth.UsageMiddleware(recorder, nil),
	})

	return httptest.NewServer(handler)
}

// --- Tests ---

func TestAuth_ValidKey(t *testing.T) {
	rawKey := "test-valid-key"
	keyHash := auth.HashKey(rawKey)

	lookup := &inMemoryKeyLookup{
		keys: map[string]*auth.APIKey{
			keyHash: {ID: "k1", TenantID: "t1", Status: auth.KeyStatusActive},
		},
	}
	limiter := &inMemoryRateLimiter{allowed: true}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/geocode/forward", strings.NewReader(`{"query":"test"}`))
	req.Header.Set("X-API-Key", rawKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuth_InvalidKey(t *testing.T) {
	lookup := &inMemoryKeyLookup{keys: map[string]*auth.APIKey{}}
	limiter := &inMemoryRateLimiter{allowed: true}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/geocode/forward", strings.NewReader(`{"query":"test"}`))
	req.Header.Set("X-API-Key", "bad-key")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_RevokedKey(t *testing.T) {
	rawKey := "revoked-key"
	keyHash := auth.HashKey(rawKey)

	lookup := &inMemoryKeyLookup{
		keys: map[string]*auth.APIKey{
			keyHash: {ID: "k2", TenantID: "t1", Status: auth.KeyStatusRevoked},
		},
	}
	limiter := &inMemoryRateLimiter{allowed: true}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/geocode/forward", strings.NewReader(`{"query":"test"}`))
	req.Header.Set("X-API-Key", rawKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_RateLimited(t *testing.T) {
	rawKey := "rate-limited-key"
	keyHash := auth.HashKey(rawKey)

	lookup := &inMemoryKeyLookup{
		keys: map[string]*auth.APIKey{
			keyHash: {ID: "k3", TenantID: "t1", Status: auth.KeyStatusActive},
		},
	}
	limiter := &inMemoryRateLimiter{allowed: false}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/geocode/forward", strings.NewReader(`{"query":"test"}`))
	req.Header.Set("X-API-Key", rawKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
}

func TestAuth_UsageRecorded(t *testing.T) {
	rawKey := "usage-key"
	keyHash := auth.HashKey(rawKey)

	lookup := &inMemoryKeyLookup{
		keys: map[string]*auth.APIKey{
			keyHash: {ID: "k4", TenantID: "t-usage", Status: auth.KeyStatusActive},
		},
	}
	limiter := &inMemoryRateLimiter{allowed: true}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/geocode/forward", strings.NewReader(`{"query":"test"}`))
	req.Header.Set("X-API-Key", rawKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Usage recording is async, give it a moment
	time.Sleep(100 * time.Millisecond)

	if len(recorder.records) == 0 {
		t.Fatal("expected at least one usage record")
	}

	rec := recorder.records[0]
	if rec.TenantID != "t-usage" {
		t.Fatalf("expected tenant t-usage, got %s", rec.TenantID)
	}

	if rec.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.StatusCode)
	}
}

func TestAuth_MissingKey(t *testing.T) {
	lookup := &inMemoryKeyLookup{keys: map[string]*auth.APIKey{}}
	limiter := &inMemoryRateLimiter{allowed: true}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/geocode/forward", strings.NewReader(`{"query":"test"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	var errResp api.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error: %v", err)
	}

	if errResp.Error.Code != "unauthorized" {
		t.Fatalf("expected unauthorized code, got %s", errResp.Error.Code)
	}
}

func TestAuth_HealthEndpointNoAuth(t *testing.T) {
	lookup := &inMemoryKeyLookup{keys: map[string]*auth.APIKey{}}
	limiter := &inMemoryRateLimiter{allowed: true}
	recorder := &inMemoryUsageRecorder{}

	srv := buildTestServer(lookup, limiter, recorder)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
