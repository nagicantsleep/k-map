# Release Checklist — k-map Beta

Use this checklist before handing the service off to the first beta tenants. All items must be checked before the epic is closed.

---

## 1. Quality Gates

### 1.1 Golden Dataset Validation (Issue #43)
- [ ] `go test -v -tags integration ./test/... -run TestGoldenDataset` runs against the target deployment
- [ ] All 5 forward geocoding fixture cases pass
- [ ] All 5 reverse geocoding fixture cases pass
- [ ] All 5 proximity fixture cases pass
- [ ] Test report posted to issue #43

### 1.2 Load Test (Issue #44)
- [ ] `k6 run scripts/load/k6_load_test.js` runs against the target deployment (50 VUs, 2 minutes)
- [ ] Forward geocode P95 latency < 800 ms
- [ ] Reverse geocode P95 latency < 500 ms
- [ ] Proximity check P95 latency < 1000 ms
- [ ] 5xx error rate < 1% during the load run
- [ ] k6 summary output posted to issue #44

### 1.3 Config Hardening (Issue #45)
- [ ] `KMAP_ENV=production` is set in the target deployment
- [ ] `KMAP_POSTGRES_DSN` is explicitly set to a non-default, secure DSN
- [ ] API starts cleanly without config validation errors
- [ ] `.env.example` is reviewed and up to date

### 1.4 API Documentation (Issue #46)
- [ ] `docs/api-guide.md` is complete and reviewed
- [ ] All curl examples verified against the target deployment
- [ ] Documentation is accessible to beta tenants

---

## 2. Infrastructure and Deployment

- [ ] Nominatim is running and healthy (`GET /readyz` shows `nominatim: ok`)
- [ ] PostgreSQL is running and healthy (`GET /readyz` shows `postgres: ok`)
- [ ] Redis is running and healthy (`GET /readyz` shows `redis: ok`)
- [ ] API service starts cleanly and logs show `"starting api server"`
- [ ] `GET /healthz` returns `200 OK {"status":"ok"}`
- [ ] `GET /readyz` returns `200 OK` with all dependencies healthy
- [ ] Postgres migrations ran successfully on startup (check startup logs)
- [ ] OSM dataset has been imported for the target launch region (verify via a known address lookup)

---

## 3. Security

- [ ] TLS termination is configured at the ingress/load balancer layer
- [ ] API keys are stored hashed at rest (verify via database inspection — no plaintext keys in `api_keys` table)
- [ ] Insecure default credentials (`kmap:kmap`) are not in use
- [ ] No secrets are visible in application logs (run a test request and inspect logs)
- [ ] Nominatim is not publicly accessible (only reachable from the API service via private networking)

---

## 4. Observability

- [ ] Structured JSON logs are emitting with `request_id` and `tenant_id` on every API request
- [ ] Prometheus metrics endpoint is reachable (default: `:8080/metrics`)
- [ ] At minimum the following metrics are present:
  - `kmap_requests_total`
  - `kmap_request_duration_seconds`
  - `kmap_geocoder_duration_seconds`
  - `kmap_cache_hits_total`
  - `kmap_cache_misses_total`

---

## 5. Tenant Onboarding Smoke Test

- [ ] First beta tenant record created in the database
- [ ] API key created and associated with the tenant
- [ ] Beta tenant can successfully call `POST /v1/geocode/forward` with the issued key
- [ ] Beta tenant can successfully call `POST /v1/geocode/reverse` with the issued key
- [ ] Beta tenant can successfully call `POST /v1/geocode/proximity` with the issued key
- [ ] Invalid key returns `401 unauthorized`
- [ ] Rate limit enforcement verified (exceed the per-minute limit and confirm `429` response)

---

## 6. Documentation Handoff

- [ ] `docs/api-guide.md` shared with beta tenants
- [ ] `docs/known-limitations.md` shared with beta tenants
- [ ] Beta tenants acknowledge the limitations (especially proximity point-to-point semantics)
- [ ] Support contact or escalation path communicated to beta tenants

---

## 7. Sign-Off

| Item                          | Status | Notes |
|-------------------------------|--------|-------|
| Golden dataset validation     | ☐      |       |
| Load test                     | ☐      |       |
| Config hardening              | ☐      |       |
| API documentation             | ☐      |       |
| Infrastructure health         | ☐      |       |
| Security review               | ☐      |       |
| Observability verified        | ☐      |       |
| Tenant smoke test             | ☐      |       |
| Documentation handoff         | ☐      |       |

**Release approved by:** ___________________________  
**Date:** ___________________________
