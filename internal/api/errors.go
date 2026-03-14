package api

import (
	"encoding/json"
	"net/http"
)

// Error codes for API responses.
const (
	ErrCodeValidation        = "validation_error"
	ErrCodeUnauthorized      = "unauthorized"
	ErrCodeNotFound          = "not_found"
	ErrCodeRateLimitExceeded = "rate_limit_exceeded"
	ErrCodeInternal          = "internal_error"
	ErrCodeGeocoderUnavailable = "geocoder_unavailable"
)

// ErrorDetail represents the error object in an error response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse represents the standard error response shape.
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id"`
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, statusCode int, code, message, requestID string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

// WriteBadRequest writes a 400 validation error response.
func WriteBadRequest(w http.ResponseWriter, message, requestID string) {
	WriteError(w, http.StatusBadRequest, ErrCodeValidation, message, requestID)
}

// WriteUnauthorized writes a 401 unauthorized error response.
func WriteUnauthorized(w http.ResponseWriter, message, requestID string) {
	WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, message, requestID)
}

// WriteNotFound writes a 404 not found error response.
func WriteNotFound(w http.ResponseWriter, message, requestID string) {
	WriteError(w, http.StatusNotFound, ErrCodeNotFound, message, requestID)
}

// WriteRateLimitExceeded writes a 429 rate limit error response.
func WriteRateLimitExceeded(w http.ResponseWriter, message, requestID string) {
	WriteError(w, http.StatusTooManyRequests, ErrCodeRateLimitExceeded, message, requestID)
}

// WriteInternalError writes a 500 internal error response.
func WriteInternalError(w http.ResponseWriter, requestID string) {
	WriteError(w, http.StatusInternalServerError, ErrCodeInternal, "Internal server error", requestID)
}

// WriteGeocoderUnavailable writes a 503 response for geocoder outages.
func WriteGeocoderUnavailable(w http.ResponseWriter, requestID string) {
	WriteError(w, http.StatusServiceUnavailable, ErrCodeGeocoderUnavailable, "Geocoding service temporarily unavailable", requestID)
}
