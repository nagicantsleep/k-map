# Architecture: OSM-Backed Geolocation SaaS MVP

## 1. Purpose

This document defines the target MVP architecture for `k-map`, an OSM-backed SaaS geolocation API. It translates the product requirements in `docs/PRD.md` into a minimal system design that is practical to implement and operate.

## 2. Architecture Goals

- Keep the number of moving parts low
- Support the three MVP capabilities:
  - forward geocoding
  - reverse geocoding
  - proximity validation
- Own the commercial reliability of the service by self-hosting the geocoder layer
- Preserve clear upgrade paths for scaling, data refreshes, and enterprise isolation

## 3. Non-Goals

- Map tile serving
- Route calculation
- Consumer-facing web or mobile UI
- Full-text search infrastructure beyond what the geocoder requires
- Event-driven architecture for MVP

## 4. Recommended Stack

### Application Layer

- Language: `Go`
- API style: `REST/JSON`
- Transport: `net/http` or a minimal router

### Data and Geo Layer

- Geocoder: self-hosted `Nominatim`
- Primary database: `PostgreSQL`
- Geospatial extension: `PostGIS`
- Cache and rate limiting: `Redis`

### Runtime and Deployment

- Local development: `Docker Compose`
- Production: container platform or `Kubernetes`
- Observability: structured logs, metrics, health/readiness probes

## 5. High-Level Architecture

```text
Clients
  |
  v
Public API Service (Go)
  |- Auth / API Keys
  |- Rate Limiting
  |- Request Validation
  |- Response Normalization
  |- Usage Tracking
  |- Proximity Engine
  |
  +--> Redis
  |      |- request throttling
  |      |- short-lived response cache
  |
  +--> App Postgres
  |      |- tenants
  |      |- api keys
  |      |- usage records
  |
  +--> Internal Geocoder Adapter
         |
         v
      Nominatim
         |
         v
   Postgres + PostGIS (OSM data)
```

## 6. Service Boundaries

### 6.1 Public API Service

The API service is the only public-facing runtime in MVP. It is responsible for:

- authenticating requests
- applying rate limits
- validating payloads
- calling the internal geocoder adapter
- normalizing responses into stable contracts
- computing proximity results
- emitting logs, metrics, and usage events

The API service should remain stateless so it can scale horizontally.

### 6.2 Geocoder Layer

`Nominatim` is used for forward and reverse geocoding against imported OSM data.

The API service must not expose raw Nominatim responses directly. Instead, it should use an internal adapter that:

- maps Nominatim fields to stable public DTOs
- applies result ranking and confidence mapping rules
- isolates the rest of the codebase from upstream response changes

### 6.3 Data Stores

There are two logical data concerns:

1. Application metadata
   - tenants
   - API keys
   - usage records
   - optional cached metadata
2. OSM geocoder data
   - imported and indexed by Nominatim

These can share the same Postgres cluster in early MVP, but should remain logically separated. For cleaner operations, separate databases within one cluster are preferred.

## 7. Main Request Flows

### 7.1 Forward Geocoding

1. Client calls `POST /v1/geocode/forward`
2. API validates API key and rate limit
3. API normalizes and hashes the request for cache lookup
4. If cache miss, API calls Nominatim search via internal adapter
5. Adapter maps results to public response shape
6. API stores short-lived cache entry if safe
7. API records usage and metrics
8. API returns ranked results

### 7.2 Reverse Geocoding

1. Client calls `POST /v1/geocode/reverse`
2. API validates coordinates, API key, and rate limit
3. API checks cache
4. If needed, API calls Nominatim reverse endpoint via adapter
5. Adapter maps upstream fields to normalized response
6. API records usage and metrics
7. API returns best result

### 7.3 Proximity Validation

1. Client calls `POST /v1/geocode/proximity`
2. API validates coordinate, target query, threshold, API key, and rate limit
3. API geocodes the target query through the internal adapter
4. API selects the best target candidate
5. API computes geodesic distance between input coordinate and target candidate
6. API returns:
   - `distance_meters`
   - `threshold_meters`
   - `is_near`
   - matched target metadata

## 8. Proximity Logic

### MVP Decision Rule

- Resolve the target address or place to the best single point result
- Compute point-to-point geodesic distance
- Return `is_near = true` when `distance_meters <= threshold_meters`

### Rationale

- This is easy to explain
- It is deterministic
- It avoids over-promising rooftop or parcel accuracy that OSM data may not consistently support

### Known Limitation

This does not prove the coordinate is inside a building, parcel, or venue boundary. It only proves closeness to the best resolved point.

## 9. Data Model Overview

### Application Database Tables

#### `tenants`

- `id`
- `name`
- `plan`
- `status`
- `created_at`

#### `api_keys`

- `id`
- `tenant_id`
- `key_hash`
- `status`
- `created_at`
- `last_used_at`

#### `usage_records`

- `id`
- `tenant_id`
- `endpoint`
- `request_id`
- `status_code`
- `latency_ms`
- `created_at`

#### Optional `cache_entries`

- `cache_key`
- `endpoint`
- `payload_hash`
- `response_json`
- `expires_at`

### Geocoder Database

Managed primarily by Nominatim and its import/indexing process. Do not tightly couple application migrations to Nominatim internals.

## 10. API Contract Rules

- Public API responses must be stable and independent of upstream field names
- All public endpoints except health/readiness require API key auth
- Every response should include a request ID
- Validation errors should be explicit and machine-readable
- Empty search results are valid business outcomes, not server failures

## 11. Caching Strategy

### Cache Candidates

- forward geocode requests with normalized input
- reverse geocode requests rounded to a safe coordinate precision
- repeated proximity lookups when both coordinate and target query are identical

### Cache Constraints

- Use short TTLs in MVP
- Cache only normalized response payloads
- Never let cache become the source of truth for tenant usage or auth

## 12. Security and Multi-Tenancy

### Authentication

- Use API keys for MVP
- Store keys hashed or encrypted at rest
- Support revocation and rotation

### Authorization

- Tenant scope is enforced by API key ownership
- Admin-only operations should be separated from public endpoints

### Privacy

- Avoid storing raw query payloads longer than operationally necessary
- Redact sensitive values from logs where practical

## 13. Observability

### Logs

Emit structured logs containing:

- request ID
- tenant ID
- endpoint
- latency
- status code
- upstream geocoder timing

### Metrics

Track at minimum:

- request count
- latency by endpoint
- error rate
- cache hit rate
- geocoder upstream failures
- rate limit rejections

### Health Endpoints

- `/healthz`: process health
- `/readyz`: downstream dependency readiness

## 14. Deployment Model

### Local Development

Use `Docker Compose` to run:

- API service
- Redis
- App Postgres
- Nominatim stack

### Production MVP

Recommended minimum runtime units:

- `api` deployment
- `redis` instance
- `postgres` instance or managed service
- `nominatim` deployment with attached storage

Production should isolate public API from internal geocoder access using private networking.

## 15. Scaling Strategy

### Scale Horizontally

- API service replicas
- Redis sized for rate limiting and cache throughput

### Scale Carefully

- Nominatim is data-heavy and more operationally sensitive than the API layer
- OSM import and refreshes must be planned, observable, and reversible

### Expected Bottlenecks

- geocoder query latency
- Nominatim import/index duration
- disk and memory pressure on geodata storage

## 16. Failure Modes and Handling

### Geocoder Failure

- Return `5xx` with stable error codes
- Surface upstream timeout/failure metrics
- Avoid leaking raw upstream error bodies

### Cache Failure

- Degrade to direct geocoder calls
- Do not block core API behavior

### Database Failure

- Fail closed for auth and tenant-scoped operations
- Keep error messages generic to callers

## 17. Data Refresh Strategy

- Import OSM data by target launch region first
- Automate import via scripts under `scripts/` or `deploy/`
- Document source extracts, import commands, and rollback steps
- Refresh on a scheduled cadence appropriate to target market needs

## 18. Initial Directory Ownership

Recommended ownership boundaries once implementation starts:

- `cmd/api`: process entrypoint and bootstrap
- `internal/api`: HTTP handlers, middleware, DTOs
- `internal/auth`: API keys and tenant auth
- `internal/geocode`: Nominatim adapter and response normalization
- `internal/proximity`: distance logic and threshold rules
- `internal/storage`: Postgres and Redis integrations
- `internal/telemetry`: logging, metrics, tracing hooks
- `deploy/compose`: local runtime orchestration
- `scripts/`: import and operational automation

## 19. Deferred Architecture Decisions

These are intentionally deferred until required:

- batch processing architecture
- customer-facing dashboard
- enterprise single-tenant deployments
- advanced ranking customization
- building or parcel polygon validation

## 20. Implementation Guidance

Build the system in this order:

1. Stand up local Postgres, Redis, and Nominatim
2. Build the API skeleton and health endpoints
3. Add forward and reverse geocoding endpoints
4. Add auth, rate limiting, and usage persistence
5. Add proximity logic and golden test fixtures
6. Add OSM refresh automation and production hardening
