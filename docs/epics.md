# Epics: `k-map` MVP Implementation Roadmap

## Purpose

This document defines the implementation epics required to build the `k-map` MVP in a controlled order. It is derived from:

- `docs/BRD.md`
- `docs/PRD.md`
- `docs/architecture.md`

The epics below are intentionally minimal. Each epic should produce a usable increment and keep scope constrained to the MVP:

- forward geocoding
- reverse geocoding
- proximity validation
- API key auth
- tenant rate limiting and usage tracking

## Execution Order

Recommended delivery order:

1. Epic 1: Repository and Service Foundation
2. Epic 2: Local Infrastructure and Geocoder Runtime
3. Epic 3: Core API Skeleton and Platform Middleware
4. Epic 4: Forward and Reverse Geocoding
5. Epic 5: Proximity Validation
6. Epic 6: Tenant Auth, Quotas, and Usage Tracking
7. Epic 7: Observability and Operational Hardening
8. Epic 8: Beta Readiness and Release Preparation

## Epic 1: Repository and Service Foundation

### Goal

Create the baseline codebase structure so the service can be developed consistently.

### Why First

All later work depends on clear package boundaries, config handling, and a runnable service entrypoint.

### Scope

- create Go module
- create directory structure from `AGENTS.md`
- add application bootstrap in `cmd/api`
- add config loading strategy
- add base HTTP server setup
- add health and readiness endpoints
- add linting and test commands

### Deliverables

- runnable `Go` API skeleton
- environment/config conventions
- initial make or script wrappers if needed
- CI-ready lint/test command set

### Dependencies

- none

### Exit Criteria

- `go run ./cmd/api` starts the service
- `/healthz` returns success
- `/readyz` returns deterministic readiness behavior
- `go test ./...` and `golangci-lint run` can be executed locally

## Epic 2: Local Infrastructure and Geocoder Runtime

### Goal

Stand up the minimum local runtime for application development and OSM-backed geocoding.

### Why This Matters

The hardest external dependency in this product is the geocoder/data layer. It should be stabilized early.

### Scope

- define `Docker Compose` stack
- add `PostgreSQL`
- add `Redis`
- add self-hosted `Nominatim`
- document local startup and shutdown
- document OSM import approach for the initial launch region

### Deliverables

- local infra definition under `deploy/compose`
- working local geocoder runtime
- setup notes for seed/import data

### Dependencies

- Epic 1

### Exit Criteria

- one command starts local dependencies
- API service can resolve internal connections to Postgres, Redis, and Nominatim
- at least one launch-region dataset is imported and queryable in local development

## Epic 3: Core API Skeleton and Platform Middleware

### Goal

Build the shared HTTP and service plumbing needed for all business endpoints.

### Scope

- request routing under `/v1`
- request ID middleware
- structured logging middleware
- error model and response helpers
- request validation patterns
- internal client/adapter for Nominatim
- shared response DTOs

### Deliverables

- stable API conventions
- internal geocoder adapter interface
- shared middleware stack

### Dependencies

- Epic 1
- Epic 2

### Exit Criteria

- handlers can call the Nominatim adapter through a defined interface
- all endpoints return a consistent error shape
- request logs include request ID, route, latency, and status code

## Epic 4: Forward and Reverse Geocoding

### Goal

Deliver the first core product capability: address-to-coordinate and coordinate-to-address resolution.

### Scope

- implement `POST /v1/geocode/forward`
- implement `POST /v1/geocode/reverse`
- normalize upstream Nominatim responses
- define confidence and source metadata
- add caching for safe repeated lookups
- add fixtures for known valid and invalid cases

### Deliverables

- production-shaped forward geocoding endpoint
- production-shaped reverse geocoding endpoint
- normalized response schema

### Dependencies

- Epic 3

### Exit Criteria

- both endpoints work end-to-end against local Nominatim
- no-match cases return empty results or explicit null result, not server errors
- integration tests cover valid, invalid, and no-match scenarios

## Epic 5: Proximity Validation

### Goal

Implement the near/not-near decision capability based on geocoded target points.

### Scope

- implement `POST /v1/geocode/proximity`
- geocode target query through the internal adapter
- compute geodesic distance
- apply threshold logic
- return matched target metadata and decision explanation fields
- add labeled near/not-near fixtures

### Deliverables

- proximity validation endpoint
- tested distance utility
- deterministic threshold behavior

### Dependencies

- Epic 4

### Exit Criteria

- endpoint returns `distance_meters`, `threshold_meters`, and `is_near`
- behavior is deterministic for the same request
- tests cover near, not-near, invalid target, and invalid coordinate cases

## Epic 6: Tenant Auth, Quotas, and Usage Tracking

### Goal

Make the service commercially usable as a multi-tenant SaaS.

### Scope

- API key model and storage
- API key authentication middleware
- tenant model
- per-tenant rate limiting via Redis
- per-tenant usage recording
- key rotation and revocation support

### Deliverables

- protected public endpoints
- quota enforcement
- usage data suitable for billing/reporting later

### Dependencies

- Epic 3
- Epic 4
- Epic 5

### Exit Criteria

- all non-health endpoints require valid API keys
- invalid and revoked keys are rejected
- rate-limited tenants receive `429`
- usage is recorded by tenant and endpoint

## Epic 7: Observability and Operational Hardening

### Goal

Make the MVP operable under real tenant traffic and support debugging, monitoring, and data refreshes.

### Scope

- metrics emission
- cache hit/miss measurement
- upstream geocoder timing metrics
- readiness dependency checks
- timeout and retry policy for internal geocoder calls
- log field standardization
- OSM refresh/import runbook

### Deliverables

- baseline dashboards or metric definitions
- operational runbook for geocoder refresh
- hardened dependency handling

### Dependencies

- Epic 2
- Epic 3
- Epic 4
- Epic 5
- Epic 6

### Exit Criteria

- latency, error, and upstream dependency metrics are available
- readiness reflects dependency state
- geocoder outages fail predictably with stable API errors
- OSM refresh procedure is documented and reproducible

## Epic 8: Beta Readiness and Release Preparation

### Goal

Prepare the system for first external tenant usage.

### Scope

- golden dataset validation for launch region
- load and smoke testing
- config hardening for non-local environments
- release checklist
- API usage examples and integration docs
- explicit launch constraints and known limitations

### Deliverables

- beta-ready release candidate
- operator checklist
- consumer-facing API examples

### Dependencies

- Epic 4
- Epic 5
- Epic 6
- Epic 7

### Exit Criteria

- launch-region validation dataset meets agreed quality bar
- basic load test completes without critical failures
- known limitations are documented
- service can be handed to first beta tenants

## Suggested Issue Breakdown Pattern

Each epic should be split into small issues that change one concern at a time.

Example pattern:

- bootstrap and configuration
- Docker Compose and dependency wiring
- Nominatim adapter
- forward geocode handler
- reverse geocode handler
- proximity distance engine
- API key middleware
- Redis rate limiter
- usage persistence
- metrics and dashboards

## Rules for Epic Creation

- Do not create an epic for features outside MVP scope.
- Do not combine infrastructure setup and business endpoint behavior into one issue unless the change is trivial.
- Any issue that changes public API shape must also update:
  - `docs/PRD.md`
  - `docs/architecture.md`
- Any issue that changes operational dependencies must update:
  - `docs/architecture.md`
  - `AGENTS.md` if workflow or commands change

## Recommended First Epic to Start

Start with `Epic 1: Repository and Service Foundation`.

Reason:

- it removes ambiguity from the repo structure
- it enables parallel work later
- it is the lowest-risk path to getting implementation moving
