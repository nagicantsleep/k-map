package auth

import (
	"context"
	"database/sql"
	"fmt"
)

// UsageRecorder records API usage per tenant and endpoint.
type UsageRecorder struct {
	db *sql.DB
}

// NewUsageRecorder creates a new Postgres-backed usage recorder.
func NewUsageRecorder(db *sql.DB) *UsageRecorder {
	return &UsageRecorder{db: db}
}

// Record inserts a usage record.
func (u *UsageRecorder) Record(ctx context.Context, tenantID, endpoint, requestID string, statusCode, latencyMs int) error {
	_, err := u.db.ExecContext(ctx,
		`INSERT INTO usage_records (tenant_id, endpoint, request_id, status_code, latency_ms)
		 VALUES ($1, $2, $3, $4, $5)`,
		tenantID, endpoint, requestID, statusCode, latencyMs,
	)
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}

	return nil
}
