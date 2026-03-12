package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/nagicantsleep/k-map/internal/api"
)

// UsageRecorderInterface defines the interface for recording usage.
type UsageRecorderInterface interface {
	Record(ctx context.Context, tenantID, endpoint, requestID string, statusCode, latencyMs int) error
}

// UsageMiddleware records per-tenant usage after each request completes.
func UsageMiddleware(recorder UsageRecorderInterface, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &usageResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			tenantID := TenantIDFromContext(r.Context())
			if tenantID == "" {
				return
			}

			requestID := api.RequestIDFromContext(r.Context())
			latencyMs := int(time.Since(start).Milliseconds())

			go func() {
				if err := recorder.Record(context.Background(), tenantID, r.URL.Path, requestID, rw.statusCode, latencyMs); err != nil {
					logger.Error("failed to record usage", "error", err, "tenant_id", tenantID)
				}
			}()
		})
	}
}

// usageResponseWriter captures the status code for usage recording.
type usageResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *usageResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *usageResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}

	return w.ResponseWriter.Write(b)
}
