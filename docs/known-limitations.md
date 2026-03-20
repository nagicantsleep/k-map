# Known Limitations — k-map MVP

This document records explicit constraints and known limitations of the k-map MVP. These should be communicated to all beta tenants before onboarding.

---

## 1. Proximity is Point-to-Point, Not Rooftop or Parcel Containment

The `POST /v1/geocode/proximity` endpoint computes a **geodesic distance** between the input coordinate and the best geocoded candidate point for the target address.

**What this means:**
- `is_near: true` means the input coordinate is within `threshold_meters` of the resolved candidate point.
- It does **not** prove the coordinate is inside a building, on a specific parcel, or within a venue boundary.
- The resolved candidate point may be at a street-level centroid, a postal code centroid, or a named place centroid — not necessarily a precise building entrance.

**Implication for callers:** Applications that require rooftop-level certainty or polygon containment must not rely on this endpoint for those decisions.

---

## 2. Geocoder Accuracy is Bounded by OSM Data Coverage

All geocoding is performed against OpenStreetMap (OSM) data imported for the configured launch region.

**What this means:**
- Addresses, places, or points of interest that are not present in the OSM extract will not be resolved.
- Accuracy and completeness varies by geography and contributor density.
- Structured address fields (house number, street, city) are only as reliable as the OSM tagging in the imported dataset.

**Implication for callers:** Do not assume 100% coverage or street-level accuracy for all queries, especially in areas with sparse OSM data.

---

## 3. No Batch Processing in MVP

The API processes one request at a time, synchronously. There is no batch endpoint for geocoding multiple addresses in a single request.

**Implication for callers:** High-volume address processing must be handled by calling the API in parallel, subject to per-tenant rate limits. Batch async processing is deferred to a post-MVP release.

---

## 4. No Tenant Dashboard or Self-Service Key Management

Tenant registration, API key creation, and key rotation are **admin-only operations** in MVP. There is no tenant-facing dashboard or self-service portal.

**Implication for callers:** Beta tenants must request key creation or rotation from the service operator via out-of-band communication.

---

## 5. OSM Data Refresh is Manual and Scripted

The imported OSM dataset does not update automatically. Refreshing the geocoder with newer OSM data requires running the import script manually or scheduling it as an operational task.

**Implication for callers:** Address changes, new streets, or new places in the real world may not appear in the geocoder immediately. See `scripts/import-osm.sh` and `docs/architecture.md §17` for refresh procedures.

---

## 6. No TLS in Local Development

TLS is not configured for the Docker Compose local dev stack. All traffic on `localhost:8080` is plain HTTP.

**Implication:** TLS must be terminated at the load balancer or ingress layer in all non-local environments. The API binary itself does not currently handle TLS termination.

---

## 7. Rate Limits Apply Per Tenant, Not Per Endpoint

Rate limiting is enforced per tenant (API key) across all endpoints combined, not per individual endpoint. A high rate of proximity checks will reduce the remaining quota for forward and reverse geocoding within the same minute.

---

## 8. Confidence Scores are Heuristic

The `confidence` field in geocoding responses is derived from Nominatim's internal ranking signals and mapped to a 0.0–1.0 scale. It is a heuristic indicator of match quality, not a statistically calibrated probability.

**Implication for callers:** Do not use `confidence` as a hard threshold for automated decisions without validating its behavior against your specific query patterns.
