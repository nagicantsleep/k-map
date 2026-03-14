package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nagicantsleep/k-map/internal/api"
)

// mockRateLimitChecker is a test double for RateLimitChecker.
type mockRateLimitChecker struct {
	allowed bool
	err     error
}

func (m *mockRateLimitChecker) Allow(_ context.Context, _ string) (bool, error) {
	return m.allowed, m.err
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	checker := &mockRateLimitChecker{allowed: true}
	handler := RateLimitMiddleware(checker)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(api.WithTenantID(req.Context(), "tenant-1"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_RateLimited(t *testing.T) {
	checker := &mockRateLimitChecker{allowed: false}
	handler := RateLimitMiddleware(checker)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(api.WithTenantID(req.Context(), "tenant-1"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_MissingTenant(t *testing.T) {
	checker := &mockRateLimitChecker{allowed: true}
	handler := RateLimitMiddleware(checker)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
