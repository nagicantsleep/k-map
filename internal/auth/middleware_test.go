package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockKeyLookup is a test double for KeyLookup.
type mockKeyLookup struct {
	key *APIKey
	err error
}

func (m *mockKeyLookup) LookupByHash(_ context.Context, _ string) (*APIKey, error) {
	return m.key, m.err
}

func (m *mockKeyLookup) TouchAPIKey(_ context.Context, _ string) error {
	return nil
}

func TestAuthMiddleware_MissingKey(t *testing.T) {
	handler := AuthMiddleware(&mockKeyLookup{})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/geocode/forward", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidKey(t *testing.T) {
	lookup := &mockKeyLookup{err: sql.ErrNoRows}
	handler := AuthMiddleware(lookup)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "bad-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_RevokedKey(t *testing.T) {
	lookup := &mockKeyLookup{
		key: &APIKey{ID: "k1", TenantID: "t1", Status: KeyStatusRevoked},
	}
	handler := AuthMiddleware(lookup)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "some-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidKey(t *testing.T) {
	lookup := &mockKeyLookup{
		key: &APIKey{ID: "k1", TenantID: "tenant-123", Status: KeyStatusActive},
	}

	var capturedTenantID string
	handler := AuthMiddleware(lookup)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = TenantIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "valid-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if capturedTenantID != "tenant-123" {
		t.Fatalf("expected tenant-123, got %s", capturedTenantID)
	}
}
