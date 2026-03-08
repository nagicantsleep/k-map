package api

import "net/http"

// NewHandler builds the base handler graph for the public API.
func NewHandler() http.Handler {
	return http.NewServeMux()
}
