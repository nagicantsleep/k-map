CREATE TABLE usage_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    endpoint    TEXT NOT NULL,
    request_id  TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    latency_ms  INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_usage_records_tenant_id ON usage_records(tenant_id);
CREATE INDEX idx_usage_records_created_at ON usage_records(created_at);
