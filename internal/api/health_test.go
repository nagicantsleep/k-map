package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandlerHealthz(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	NewHandler(HandlerOptions{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}

	assertStatusResponse(t, recorder, "ok")
}

func TestNewHandlerReadyz(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	NewHandler(HandlerOptions{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}

	assertStatusResponse(t, recorder, "ok")
}

func TestNewHandlerRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)

	NewHandler(HandlerOptions{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func TestNewHandlerReadyzReturnsServiceUnavailableWhenDependenciesFail(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	NewHandler(HandlerOptions{
		ReadinessChecker: stubReadinessChecker{err: context.DeadlineExceeded},
	}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}

	assertStatusResponse(t, recorder, "degraded")
}

func assertStatusResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus string) {
	t.Helper()

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var response statusResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Status != wantStatus {
		t.Fatalf("Status = %q, want %q", response.Status, wantStatus)
	}
}

type stubReadinessChecker struct {
	err error
}

func (checker stubReadinessChecker) Check(_ context.Context) error {
	return checker.err
}

func (checker stubReadinessChecker) CheckAll(_ context.Context) ReadinessResult {
	if checker.err != nil {
		return ReadinessResult{
			Status:       "degraded",
			Dependencies: map[string]string{"test": checker.err.Error()},
		}
	}
	return ReadinessResult{
		Status:       "ok",
		Dependencies: map[string]string{"test": "ok"},
	}
}
