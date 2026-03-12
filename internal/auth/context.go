package auth

import "context"

const tenantIDKey contextKey = "tenantID"

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// TenantIDFromContext retrieves the tenant ID from the request context.
func TenantIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(tenantIDKey).(string); ok {
		return id
	}

	return ""
}

// withTenantID stores the tenant ID in the context.
func withTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}
