package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/nagicantsleep/k-map/internal/api"
)

const apiKeyHeader = "X-API-Key"

// KeyLookup defines the interface for looking up API keys.
type KeyLookup interface {
	LookupByHash(ctx context.Context, keyHash string) (*APIKey, error)
	TouchAPIKey(ctx context.Context, keyID string) error
}

// AuthMiddleware validates the X-API-Key header and injects tenant ID into context.
func AuthMiddleware(lookup KeyLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := api.RequestIDFromContext(r.Context())

			rawKey := r.Header.Get(apiKeyHeader)
			if rawKey == "" {
				api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Missing API key", requestID)
				return
			}

			keyHash := HashKey(rawKey)

			apiKey, err := lookup.LookupByHash(r.Context(), keyHash)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Invalid API key", requestID)
					return
				}

				api.WriteInternalError(w, requestID)
				return
			}

			if apiKey.Status == KeyStatusRevoked {
				api.WriteError(w, http.StatusUnauthorized, "unauthorized", "API key has been revoked", requestID)
				return
			}

			// Update last_used_at asynchronously (best-effort)
			go func() {
				_ = lookup.TouchAPIKey(context.Background(), apiKey.ID)
			}()

			ctx := api.WithTenantID(r.Context(), apiKey.TenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
