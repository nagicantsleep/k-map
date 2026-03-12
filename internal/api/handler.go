package api

import (
	"log/slog"
	"net/http"
)

// HandlerOptions defines injectable transport dependencies for the base handler graph.
type HandlerOptions struct {
	Logger            *slog.Logger
	ReadinessChecker  ReadinessChecker
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

	// Placeholder for future geocode endpoints
	// These will be added in Epic 4:
	// - POST /geocode/forward
	// - POST /geocode/reverse
	// - POST /geocode/proximity

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
