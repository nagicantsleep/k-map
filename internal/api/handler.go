package api

import "net/http"

// NewHandler builds the base handler graph for the public API.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler)

	return mux
}
