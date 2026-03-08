package api

import "net/http"

// HandlerOptions defines injectable transport dependencies for the base handler graph.
type HandlerOptions struct {
	ReadinessChecker ReadinessChecker
}

// NewHandler builds the base handler graph for the public API.
func NewHandler(options HandlerOptions) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler(options.ReadinessChecker))

	return mux
}
