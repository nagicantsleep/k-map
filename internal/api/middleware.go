package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

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

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}

	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}

	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs request start and completion with structured logging.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := RequestIDFromContext(r.Context())

			logger.Debug("request started",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", requestID,
			)

			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			latency := time.Since(start)

			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"latency_ms", latency.Milliseconds(),
				"request_id", requestID,
			)
		})
	}
}
