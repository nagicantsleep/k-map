package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
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

func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := RequestIDMiddleware(LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})))

	req := httptest.NewRequest(http.MethodPost, "/v1/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	// Verify logs contain required fields
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected log output")
	}

	// Parse log lines
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	var startLog, endLog map[string]interface{}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if msg, ok := entry["msg"].(string); ok {
			if msg == "request started" {
				startLog = entry
			} else if msg == "request completed" {
				endLog = entry
			}
		}
	}

	// Verify start log
	if startLog == nil {
		t.Error("expected 'request started' log entry")
	} else {
		if startLog["method"] != "POST" {
			t.Errorf("expected method POST in start log, got %v", startLog["method"])
		}
		if startLog["path"] != "/v1/test" {
			t.Errorf("expected path /v1/test in start log, got %v", startLog["path"])
		}
	}

	// Verify end log
	if endLog == nil {
		t.Error("expected 'request completed' log entry")
	} else {
		if endLog["method"] != "POST" {
			t.Errorf("expected method POST in end log, got %v", endLog["method"])
		}
		if endLog["path"] != "/v1/test" {
			t.Errorf("expected path /v1/test in end log, got %v", endLog["path"])
		}
		if endLog["status"] != float64(http.StatusCreated) {
			t.Errorf("expected status %d in end log, got %v", http.StatusCreated, endLog["status"])
		}
		if _, ok := endLog["latency_ms"]; !ok {
			t.Error("expected latency_ms in end log")
		}
		if _, ok := endLog["request_id"]; !ok {
			t.Error("expected request_id in end log")
		}
	}
}

func TestResponseWriter_CapturesStatus(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	rw.WriteHeader(http.StatusNotFound)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rw.statusCode)
	}
}

func TestResponseWriter_DefaultsToOK(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	_, _ = rw.Write([]byte("test"))

	if rw.statusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rw.statusCode)
	}
}
