package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, ErrCodeValidation, "Invalid input", "req-123")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != ErrCodeValidation {
		t.Errorf("expected error code %s, got %s", ErrCodeValidation, resp.Error.Code)
	}

	if resp.Error.Message != "Invalid input" {
		t.Errorf("expected error message 'Invalid input', got %s", resp.Error.Message)
	}

	if resp.RequestID != "req-123" {
		t.Errorf("expected request_id 'req-123', got %s", resp.RequestID)
	}
}

func TestWriteBadRequest(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteBadRequest(rec, "Missing required field", "req-456")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != ErrCodeValidation {
		t.Errorf("expected error code %s, got %s", ErrCodeValidation, resp.Error.Code)
	}
}

func TestWriteUnauthorized(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteUnauthorized(rec, "Invalid API key", "req-789")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != ErrCodeUnauthorized {
		t.Errorf("expected error code %s, got %s", ErrCodeUnauthorized, resp.Error.Code)
	}
}

func TestWriteNotFound(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteNotFound(rec, "Resource not found", "req-abc")

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != ErrCodeNotFound {
		t.Errorf("expected error code %s, got %s", ErrCodeNotFound, resp.Error.Code)
	}
}

func TestWriteRateLimitExceeded(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteRateLimitExceeded(rec, "Rate limit exceeded", "req-def")

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != ErrCodeRateLimitExceeded {
		t.Errorf("expected error code %s, got %s", ErrCodeRateLimitExceeded, resp.Error.Code)
	}
}

func TestWriteInternalError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteInternalError(rec, "req-ghi")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != ErrCodeInternal {
		t.Errorf("expected error code %s, got %s", ErrCodeInternal, resp.Error.Code)
	}

	if resp.Error.Message != "Internal server error" {
		t.Errorf("expected generic error message, got %s", resp.Error.Message)
	}
}

func TestErrorResponse_JSONSerialization(t *testing.T) {
	t.Parallel()

	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeValidation,
			Message: "Test message",
		},
		RequestID: "req-json",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"error":{"code":"validation_error","message":"Test message"},"request_id":"req-json"}`
	if string(data) != expected {
		t.Errorf("expected JSON %s, got %s", expected, string(data))
	}
}
