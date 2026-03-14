package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nagicantsleep/k-map/internal/telemetry"
)

// RequestIDHeader is the header key for the request ID.
const RequestIDHeader = "X-Request-ID"

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// requestIDKey is the context key for the request ID.
const requestIDKey contextKey = "requestID"

// tenantIDKey is the context key for the tenant ID.
const tenantIDKey contextKey = "tenantID"

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

// TenantIDFromContext retrieves the tenant ID from the context.
func TenantIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(tenantIDKey).(string); ok {
		return id
	}

	return ""
}

// WithTenantID stores the tenant ID in the context and returns the new context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
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
				telemetry.FieldMethod, r.Method,
				telemetry.FieldPath, r.URL.Path,
				telemetry.FieldRequestID, requestID,
			)

			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			latency := time.Since(start)
			tenantID := TenantIDFromContext(r.Context())

			logger.Info("request completed",
				telemetry.FieldMethod, r.Method,
				telemetry.FieldPath, r.URL.Path,
				telemetry.FieldStatus, rw.statusCode,
				telemetry.FieldLatencyMs, latency.Milliseconds(),
				telemetry.FieldRequestID, requestID,
				telemetry.FieldTenantID, tenantID,
			)
		})
	}
}

// MetricsMiddleware records per-request Prometheus metrics (total count, latency, rate-limit rejections).
func MetricsMiddleware(m *telemetry.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			endpoint := r.URL.Path
			method := r.Method
			statusCode := fmt.Sprintf("%d", rw.statusCode)

			m.RequestTotal.WithLabelValues(endpoint, method, statusCode).Inc()
			m.RequestDuration.WithLabelValues(endpoint, method).Observe(duration)

			if rw.statusCode == http.StatusTooManyRequests {
				m.RateLimitRejections.WithLabelValues(endpoint).Inc()
			}
		})
	}
}
