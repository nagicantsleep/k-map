# PRD: OSM-Backed Geolocation SaaS MVP

## 1. Document Control

- Product: `k-map`
- Version: `0.1`
- Status: Draft
- Date: `2026-03-09`
- Related: `docs/BRD.md`

## 2. Product Summary

`k-map` is a backend geolocation API platform backed by OpenStreetMap data. The MVP provides:

1. Reverse geocoding: `lat/lng -> address`
2. Forward geocoding: `address -> lat/lng`
3. Proximity validation: `is this lat/lng near this address/place?`

The MVP optimizes for simplicity, operational control, and clear API semantics rather than maximum feature breadth.

## 3. Product Principles

- Minimal scope, strong correctness
- Clear API contracts over clever heuristics
- Self-host core data services for commercial reliability
- Prefer proven OSS components over custom geo search logic
- Keep stateless application services simple

## 4. Users and Use Cases

### Primary Users

- Backend engineers integrating geolocation APIs
- Product teams validating customer-submitted locations
- Operations systems resolving addresses and coordinates

### MVP Use Cases

- Convert a typed delivery address into coordinates
- Convert device coordinates into a readable address
- Validate whether a driver or customer is near an expected location
- Normalize addresses before storing them in an application database

## 5. Scope

### In Scope

- REST API for forward geocoding
- REST API for reverse geocoding
- REST API for proximity validation
- API key authentication
- Tenant-aware rate limiting and quotas
- Request logging and metrics
- Confidence and match metadata in responses

### Out of Scope

- Autocomplete
- Batch async processing
- Interactive maps UI
- Routing and ETA
- Polygon geofencing management
- Enterprise billing portal in MVP

## 6. Functional Requirements

### FR-1 Forward Geocoding

The system must:

- Accept an address or place query as free text
- Optionally accept structured address fields when provided
- Return ranked candidate matches
- Return normalized address fields when available
- Return coordinates and confidence metadata

Minimum response fields:

- `query`
- `results[]`
- `results[].formatted_address`
- `results[].latitude`
- `results[].longitude`
- `results[].confidence`
- `results[].source`
- `results[].components`

### FR-2 Reverse Geocoding

The system must:

- Accept latitude and longitude
- Return the best matching address or nearby place
- Return normalized address components when available
- Return a match type and confidence

Minimum response fields:

- `latitude`
- `longitude`
- `result.formatted_address`
- `result.components`
- `result.place_type`
- `result.confidence`
- `result.source`

### FR-3 Proximity Validation

The system must:

- Accept a latitude, longitude, and target address or place query
- Geocode the target address or place
- Compute the distance between the input coordinate and the best target match
- Return whether the input coordinate is within a threshold

Minimum request fields:

- `latitude`
- `longitude`
- `target_query`
- `threshold_meters`

Minimum response fields:

- `is_near`
- `distance_meters`
- `threshold_meters`
- `target_match.formatted_address`
- `target_match.latitude`
- `target_match.longitude`
- `target_match.confidence`

### FR-4 Authentication

The system must:

- Require an API key for all non-health endpoints
- Associate each API key with a tenant
- Support key rotation

### FR-5 Quotas and Rate Limiting

The system must:

- Enforce per-tenant request quotas
- Enforce per-minute rate limiting
- Return clear limit errors

### FR-6 Observability

The system must:

- Record request count, latency, error rate, and status code metrics
- Emit structured logs with request ID and tenant ID
- Track upstream geocoder timings

## 7. Non-Functional Requirements

### Reliability

- Target availability: `99.9%` for shared SaaS after MVP hardening
- Graceful failure when upstream geocoder dependencies degrade

### Performance

- P95 forward geocode latency: `< 800 ms`
- P95 reverse geocode latency: `< 500 ms`
- P95 proximity check latency: `< 1,000 ms`

These targets assume warm caches and a regionally close deployment.

### Security

- TLS required in all environments except local development
- API keys stored hashed or encrypted at rest
- Audit access to admin endpoints

### Privacy

- Do not store raw customer queries longer than necessary for operations and billing
- Support configurable log retention

### Operability

- One-command local dev startup via containers
- Automated OSM import and refresh runbook
- Health and readiness endpoints

## 8. Product Decisions

### Recommended Core Stack

- API service: `Go`
- Database: `PostgreSQL + PostGIS`
- Geocoder engine: `Nominatim` with self-hosted OSM data
- Cache and rate limiting: `Redis`
- Deployment: `Docker Compose` for local, `Kubernetes` or a small container platform for production

### Why This Stack

- `Nominatim` is the most direct fit for forward and reverse geocoding from OSM data
- `PostgreSQL/PostGIS` is proven for geo workloads and required by common OSM tooling
- `Go` keeps the API layer small, fast, and easy to operate
- `Redis` provides simple rate limiting and cache support without over-design

### Rejected MVP Alternatives

- Public Nominatim endpoints: not suitable for commercial SaaS reliability or policy reasons
- Building a custom geocoder from raw OSM immediately: unnecessary complexity
- Elasticsearch-heavy search stack at MVP: too much operational weight for the first release

## 9. System Design

### High-Level Architecture

1. Client sends request to `k-map` API
2. API authenticates tenant and applies rate limits
3. API checks cache for deterministic repeat queries
4. API calls internal geocoding service backed by self-hosted `Nominatim`
5. API normalizes response shape
6. For proximity checks, API computes geodesic distance using the best target match
7. API logs request metadata and returns response

### Minimal Services

- `api`: public REST service
- `geocoder`: internal Nominatim service
- `postgres`: geodata and app metadata
- `redis`: cache and rate limiting

### Distance Logic for Proximity

MVP rule:

- Geocode the target query to the best single candidate
- Compute geodesic distance from input coordinate to candidate coordinate
- Return `is_near = true` when `distance_meters <= threshold_meters`

Default threshold:

- `100 meters` if caller does not specify one

Important limitation:

- This is proximity-to-best-match-point, not full parcel or rooftop containment

## 10. API Draft

### `POST /v1/geocode/forward`

Request:

```json
{
  "query": "1600 Amphitheatre Parkway, Mountain View, CA",
  "limit": 5
}
```

### `POST /v1/geocode/reverse`

Request:

```json
{
  "latitude": 37.422,
  "longitude": -122.084
}
```

### `POST /v1/geocode/proximity`

Request:

```json
{
  "latitude": 37.42195,
  "longitude": -122.08405,
  "target_query": "1600 Amphitheatre Parkway, Mountain View, CA",
  "threshold_meters": 100
}
```

### Error Model

All endpoints should return:

- `400` for invalid input
- `401` for invalid API key
- `429` for rate limit exceeded
- `5xx` for internal or upstream failures

Error response shape:

```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded"
  },
  "request_id": "req_123"
}
```

## 11. Data Model

### Tenant

- `id`
- `name`
- `plan`
- `status`
- `created_at`

### API Key

- `id`
- `tenant_id`
- `key_hash`
- `status`
- `created_at`
- `last_used_at`

### Usage Record

- `id`
- `tenant_id`
- `endpoint`
- `request_count`
- `date_bucket`

### Optional Cached Result

- `cache_key`
- `endpoint`
- `payload_hash`
- `response_json`
- `expires_at`

## 12. Acceptance Criteria

### Forward Geocoding

- Returns at least one ranked result for supported valid addresses
- Returns normalized address fields when source data exists
- Returns empty results, not server error, when no match is found

### Reverse Geocoding

- Returns the closest usable address or place for valid coordinates
- Rejects invalid coordinates outside geographic bounds

### Proximity Validation

- Returns deterministic `distance_meters`
- Returns `is_near = true` only when within threshold
- Includes the matched target used for the decision

### Platform

- All API endpoints require valid API key
- Rate limiting and quotas are enforced
- Structured logs and metrics are emitted for every request

## 13. Testing Strategy

### Unit Tests

- Distance calculation
- Request validation
- Response normalization
- Rate limit behavior

### Integration Tests

- API to Nominatim request flow
- Postgres and Redis connectivity
- Auth and quota enforcement

### Dataset Validation

- Golden test set of known addresses and coordinates by target launch region
- Labeled near/not-near cases for proximity logic

## 14. Rollout Plan

### Milestone 1

- Local dev stack
- Self-hosted Nominatim import for one launch region
- Basic forward and reverse geocoding endpoints

### Milestone 2

- Proximity validation endpoint
- Auth, rate limiting, usage tracking
- Metrics and dashboards

### Milestone 3

- Beta tenants
- Load testing
- Data refresh automation

## 15. Open Questions

- Which launch geography should be optimized first
- Whether structured address input is needed in MVP or immediately after
- Whether tenant-level custom thresholds should be supported from day one
- Whether billing is externalized first or implemented in-product later

## 16. Recommended Next Build Order

1. Stand up self-hosted `Nominatim` with one region dataset
2. Build the `Go` API wrapper and normalize responses
3. Add tenant auth, rate limiting, and request logging
4. Add proximity endpoint and labeled test cases
5. Add operational tooling for OSM refreshes and dashboards
