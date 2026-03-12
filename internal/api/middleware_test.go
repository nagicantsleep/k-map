package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	t.Parallel()

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		if requestID == "" {
			t.Error("expected request ID in context, got empty string")
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	requestID := rec.Header().Get(RequestIDHeader)
	if requestID == "" {
		t.Error("expected X-Request-ID header in response")
	}
}

func TestRequestIDMiddleware_PropagatesExistingID(t *testing.T) {
	t.Parallel()

	existingID := "existing-request-id-123"

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		if requestID != existingID {
			t.Errorf("expected request ID %q in context, got %q", existingID, requestID)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	requestID := rec.Header().Get(RequestIDHeader)
	if requestID != existingID {
		t.Errorf("expected X-Request-ID header %q, got %q", existingID, requestID)
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	t.Parallel()

	requestID := RequestIDFromContext(context.Background())
	if requestID != "" {
		t.Errorf("expected empty string for empty context, got %q", requestID)
	}
}
