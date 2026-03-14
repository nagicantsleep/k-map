package auth

import (
	"context"
	"net/http"

	"github.com/nagicantsleep/k-map/internal/api"
)

// RateLimitChecker defines the interface for rate limit checks.
type RateLimitChecker interface {
	Allow(ctx context.Context, tenantID string) (bool, error)
}

// RateLimitMiddleware enforces per-tenant rate limiting.
// Must be applied after AuthMiddleware so tenant ID is in context.
func RateLimitMiddleware(checker RateLimitChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := api.RequestIDFromContext(r.Context())
			tenantID := api.TenantIDFromContext(r.Context())

			if tenantID == "" {
				api.WriteInternalError(w, requestID)
				return
			}

			allowed, err := checker.Allow(r.Context(), tenantID)
			if err != nil {
				api.WriteInternalError(w, requestID)
				return
			}

			if !allowed {
				api.WriteRateLimitExceeded(w, "Rate limit exceeded", requestID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
