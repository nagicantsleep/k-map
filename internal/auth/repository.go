package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository provides tenant and API key persistence operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new auth repository backed by Postgres.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateTenant inserts a new tenant and returns it with its generated ID.
func (r *Repository) CreateTenant(ctx context.Context, name, plan string) (*Tenant, error) {
	t := &Tenant{
		Name:   name,
		Plan:   plan,
		Status: TenantStatusActive,
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO tenants (name, plan, status) VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		t.Name, t.Plan, t.Status,
	).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return t, nil
}

// GetTenant retrieves a tenant by ID.
func (r *Repository) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	t := &Tenant{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, plan, status, created_at FROM tenants WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.Plan, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}

	return t, nil
}

// CreateAPIKey inserts a new API key record for the given tenant.
// The caller is responsible for hashing the raw key before calling this.
func (r *Repository) CreateAPIKey(ctx context.Context, tenantID, keyHash string) (*APIKey, error) {
	k := &APIKey{
		TenantID: tenantID,
		KeyHash:  keyHash,
		Status:   KeyStatusActive,
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO api_keys (tenant_id, key_hash, status) VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		k.TenantID, k.KeyHash, k.Status,
	).Scan(&k.ID, &k.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return k, nil
}

// LookupByHash finds an active API key by its hash and returns the key with tenant info.
func (r *Repository) LookupByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	k := &APIKey{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, key_hash, status, created_at, last_used_at
		 FROM api_keys WHERE key_hash = $1`,
		keyHash,
	).Scan(&k.ID, &k.TenantID, &k.KeyHash, &k.Status, &k.CreatedAt, &k.LastUsedAt)
	if err != nil {
		return nil, fmt.Errorf("lookup api key: %w", err)
	}

	return k, nil
}

// RevokeAPIKey sets the status of an API key to revoked.
func (r *Repository) RevokeAPIKey(ctx context.Context, keyID string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE api_keys SET status = $1 WHERE id = $2`,
		KeyStatusRevoked, keyID,
	)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke api key rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("revoke api key: key not found")
	}

	return nil
}

// TouchAPIKey updates the last_used_at timestamp for an API key.
func (r *Repository) TouchAPIKey(ctx context.Context, keyID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = $1 WHERE id = $2`,
		time.Now().UTC(), keyID,
	)
	if err != nil {
		return fmt.Errorf("touch api key: %w", err)
	}

	return nil
}
