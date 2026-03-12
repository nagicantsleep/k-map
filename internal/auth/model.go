package auth

import "time"

// TenantStatus represents the status of a tenant.
type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusInactive TenantStatus = "inactive"
)

// KeyStatus represents the status of an API key.
type KeyStatus string

const (
	KeyStatusActive  KeyStatus = "active"
	KeyStatusRevoked KeyStatus = "revoked"
)

// Tenant represents a registered tenant in the system.
type Tenant struct {
	ID        string
	Name      string
	Plan      string
	Status    TenantStatus
	CreatedAt time.Time
}

// APIKey represents a hashed API key associated with a tenant.
type APIKey struct {
	ID         string
	TenantID   string
	KeyHash    string
	Status     KeyStatus
	CreatedAt  time.Time
	LastUsedAt *time.Time
}
