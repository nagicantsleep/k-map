package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is the header key for the request ID.
const RequestIDHeader = "X-Request-ID"

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// requestIDKey is the context key for the request ID.
const requestIDKey contextKey = "requestID"

// RequestIDFromContext retrieves the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}

	return ""
}

// RequestIDMiddleware generates or propagates a unique request ID for each request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context for downstream handlers
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add to response headers
		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r)
	})
}
