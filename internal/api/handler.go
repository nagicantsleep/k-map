package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// Geocoder defines the interface for geocoding operations used by handlers.
type Geocoder interface {
	Search(ctx context.Context, query string, limit int) ([]GeocodeResult, error)
	Reverse(ctx context.Context, lat, lng float64) (*GeocodeResult, error)
}

// HandlerOptions defines injectable transport dependencies for the base handler graph.
type HandlerOptions struct {
	Logger           *slog.Logger
	ReadinessChecker ReadinessChecker
	Geocoder         Geocoder
}

// NewHandler builds the base handler graph for the public API.
func NewHandler(options HandlerOptions) http.Handler {
	mux := http.NewServeMux()

	// Health and readiness endpoints at root level (no auth required)
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler(options.ReadinessChecker))

	// Create v1 sub-router with middleware chain
	v1Handler := newV1Handler(options)
	mux.Handle("/v1/", http.StripPrefix("/v1", v1Handler))

	return mux
}

// newV1Handler creates the v1 API handler with middleware applied.
func newV1Handler(options HandlerOptions) http.Handler {
	mux := http.NewServeMux()

	if options.Geocoder != nil {
		mux.HandleFunc("/geocode/forward", forwardGeocodeHandler(options.Geocoder))
		mux.HandleFunc("/geocode/reverse", reverseGeocodeHandler(options.Geocoder))
	}

	// Build middleware chain
	var handler http.Handler = mux

	// Apply logging middleware (requires request ID from RequestIDMiddleware)
	if options.Logger != nil {
		handler = LoggingMiddleware(options.Logger)(handler)
	}

	// Apply request ID middleware (must be first in chain)
	handler = RequestIDMiddleware(handler)

	return handler
}

// forwardGeocodeHandler handles POST /v1/geocode/forward.
func forwardGeocodeHandler(geocoder Geocoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())

		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST is allowed", requestID)
			return
		}

		var req ForwardGeocodeRequest
		if err := DecodeJSON(r, &req); err != nil {
			WriteBadRequest(w, err.Error(), requestID)
			return
		}

		if err := ValidateRequiredString(req.Query, "query"); err != nil {
			WriteBadRequest(w, err.Error(), requestID)
			return
		}

		limit := req.Limit
		if limit <= 0 {
			limit = 5
		}

		results, err := geocoder.Search(r.Context(), req.Query, limit)
		if err != nil {
			WriteInternalError(w, requestID)
			return
		}

		if results == nil {
			results = []GeocodeResult{}
		}

		resp := ForwardGeocodeResponse{
			Query:   req.Query,
			Results: results,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// reverseGeocodeHandler handles POST /v1/geocode/reverse.
func reverseGeocodeHandler(geocoder Geocoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())

		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST is allowed", requestID)
			return
		}

		var req ReverseGeocodeRequest
		if err := DecodeJSON(r, &req); err != nil {
			WriteBadRequest(w, err.Error(), requestID)
			return
		}

		if err := ValidateCoordinate(req.Latitude, req.Longitude); err != nil {
			WriteBadRequest(w, err.Error(), requestID)
			return
		}

		result, err := geocoder.Reverse(r.Context(), req.Latitude, req.Longitude)
		if err != nil {
			WriteInternalError(w, requestID)
			return
		}

		resp := ReverseGeocodeResponse{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Result:    result,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
