# k-map API Integration Guide

This guide covers everything a backend engineer needs to integrate with the k-map geolocation API.

## Base URL

```
http://<host>:8080
```

Replace `<host>` with your deployment's hostname or IP. For local development using Docker Compose, the base URL is `http://localhost:8080`.

## Authentication

All API endpoints (except `/healthz` and `/readyz`) require an API key passed via the `X-API-Key` HTTP header.

```
X-API-Key: your-api-key-here
```

### Missing or invalid key responses

**Missing key (no header):**
```json
HTTP/1.1 401 Unauthorized
{
  "error": {
    "code": "unauthorized",
    "message": "missing or invalid API key"
  },
  "request_id": "req_01HX..."
}
```

**Invalid or revoked key:**
```json
HTTP/1.1 401 Unauthorized
{
  "error": {
    "code": "unauthorized",
    "message": "missing or invalid API key"
  },
  "request_id": "req_01HX..."
}
```

---

## Endpoints

### POST /v1/geocode/forward

Resolves a free-text address or place query to geographic coordinates.

#### Request

```
POST /v1/geocode/forward
Content-Type: application/json
X-API-Key: your-api-key-here
```

| Field   | Type    | Required | Description                                              |
|---------|---------|----------|----------------------------------------------------------|
| `query` | string  | yes      | Free-text address or place name to geocode               |
| `limit` | integer | no       | Maximum number of results to return (default: 5, max: 10)|

#### Request example

```bash
curl -s -X POST http://localhost:8080/v1/geocode/forward \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"query": "Mairie de Monaco", "limit": 3}'
```

#### Response

```json
HTTP/1.1 200 OK
{
  "query": "Mairie de Monaco",
  "results": [
    {
      "formatted_address": "Mairie de Monaco, 7, Rue Louis Notari, Monte-Carlo, Monaco",
      "latitude": 43.7311,
      "longitude": 7.4197,
      "confidence": 0.9,
      "source": "osm",
      "components": {
        "city": "Monaco",
        "country": "Monaco",
        "country_code": "mc"
      },
      "place_type": "amenity"
    }
  ]
}
```

| Field                         | Type    | Description                                      |
|-------------------------------|---------|--------------------------------------------------|
| `query`                       | string  | The original query echoed back                   |
| `results`                     | array   | Ranked list of matching locations (may be empty) |
| `results[].formatted_address` | string  | Human-readable full address                      |
| `results[].latitude`          | number  | WGS84 latitude                                   |
| `results[].longitude`         | number  | WGS84 longitude                                  |
| `results[].confidence`        | number  | Match confidence score (0.0–1.0)                 |
| `results[].source`            | string  | Data source (`osm`)                              |
| `results[].components`        | object  | Normalized address components                    |
| `results[].place_type`        | string  | OSM place type                                   |

**No match** — empty results array, status 200:
```json
{"query": "zzz_nonexistent_place", "results": []}
```

---

### POST /v1/geocode/reverse

Resolves geographic coordinates to the nearest address or place.

#### Request

```
POST /v1/geocode/reverse
Content-Type: application/json
X-API-Key: your-api-key-here
```

| Field       | Type   | Required | Description                        |
|-------------|--------|----------|------------------------------------|
| `latitude`  | number | yes      | WGS84 latitude (−90 to 90)         |
| `longitude` | number | yes      | WGS84 longitude (−180 to 180)      |

#### Request example

```bash
curl -s -X POST http://localhost:8080/v1/geocode/reverse \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"latitude": 43.7311, "longitude": 7.4197}'
```

#### Response

```json
HTTP/1.1 200 OK
{
  "latitude": 43.7311,
  "longitude": 7.4197,
  "result": {
    "formatted_address": "7, Rue Louis Notari, Monte-Carlo, Monaco",
    "latitude": 43.7311,
    "longitude": 7.4197,
    "confidence": 0.85,
    "source": "osm",
    "components": {
      "street_number": "7",
      "street": "Rue Louis Notari",
      "city": "Monaco",
      "country": "Monaco",
      "country_code": "mc"
    },
    "place_type": "building"
  }
}
```

| Field               | Type   | Description                                              |
|---------------------|--------|----------------------------------------------------------|
| `latitude`          | number | The input latitude                                       |
| `longitude`         | number | The input longitude                                      |
| `result`            | object | Best matching address, or `null` if no match was found   |
| `result.confidence` | number | Match confidence score (0.0–1.0)                         |

**No match** — result is null, status 200:
```json
{"latitude": 0.0, "longitude": 0.0}
```

**Invalid coordinates** — status 400:
```json
{
  "error": {"code": "invalid_request", "message": "latitude must be between -90 and 90"},
  "request_id": "req_01HX..."
}
```

---

### POST /v1/geocode/proximity

Geocodes a target address and returns whether the input coordinate is within a configurable distance threshold.

#### Request

```
POST /v1/geocode/proximity
Content-Type: application/json
X-API-Key: your-api-key-here
```

| Field              | Type   | Required | Description                                                              |
|--------------------|--------|----------|--------------------------------------------------------------------------|
| `latitude`         | number | yes      | WGS84 latitude of the point to validate (−90 to 90)                      |
| `longitude`        | number | yes      | WGS84 longitude of the point to validate (−180 to 180)                   |
| `target_query`     | string | yes      | Free-text address or place name to geocode as the target                 |
| `threshold_meters` | number | yes      | Maximum distance in meters for `is_near` to be `true` (default: 100)     |

#### Request example — near

```bash
curl -s -X POST http://localhost:8080/v1/geocode/proximity \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "latitude": 43.7311,
    "longitude": 7.4197,
    "target_query": "Mairie de Monaco",
    "threshold_meters": 500
  }'
```

#### Response — near

```json
HTTP/1.1 200 OK
{
  "is_near": true,
  "distance_meters": 45.2,
  "threshold_meters": 500,
  "target_match": {
    "formatted_address": "Mairie de Monaco, 7, Rue Louis Notari, Monte-Carlo, Monaco",
    "latitude": 43.7311,
    "longitude": 7.4197,
    "confidence": 0.9,
    "source": "osm"
  }
}
```

#### Request example — not near

```bash
curl -s -X POST http://localhost:8080/v1/geocode/proximity \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "latitude": 43.7311,
    "longitude": 7.4197,
    "target_query": "Place du Casino, Monaco",
    "threshold_meters": 100
  }'
```

#### Response — not near

```json
HTTP/1.1 200 OK
{
  "is_near": false,
  "distance_meters": 923.1,
  "threshold_meters": 100,
  "target_match": {
    "formatted_address": "Place du Casino, Monte-Carlo, Monaco",
    "latitude": 43.7393,
    "longitude": 7.4271,
    "confidence": 0.88,
    "source": "osm"
  }
}
```

| Field                         | Type    | Description                                            |
|-------------------------------|---------|--------------------------------------------------------|
| `is_near`                     | boolean | `true` if `distance_meters <= threshold_meters`        |
| `distance_meters`             | number  | Geodesic distance from input coordinate to target match |
| `threshold_meters`            | number  | The threshold that was applied                         |
| `target_match`                | object  | The geocoded target used for the proximity decision     |
| `target_match.formatted_address` | string | Best-match address for the target query             |

> **Important limitation:** Proximity is point-to-point geodesic distance to the best geocoded candidate. It does not prove the coordinate is inside a building, parcel, or venue boundary.

---

## Rate Limiting

Every API key is subject to per-minute rate limiting. When the limit is exceeded, the API returns:

```json
HTTP/1.1 429 Too Many Requests
Retry-After: 37
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "rate limit exceeded"
  },
  "request_id": "req_01HX..."
}
```

### Retry guidance

- Check the `Retry-After` header for the number of seconds to wait.
- Implement exponential backoff for programmatic retry loops.
- Default limit: 60 requests per minute per API key (configurable per tenant).

---

## Error Model

All error responses use a consistent JSON shape:

```json
{
  "error": {
    "code": "error_code",
    "message": "Human-readable description"
  },
  "request_id": "req_01HX..."
}
```

### Standard error codes

| HTTP Status | Code                  | Meaning                                             |
|-------------|-----------------------|-----------------------------------------------------|
| 400         | `invalid_request`     | Request body is malformed or missing required fields |
| 401         | `unauthorized`        | API key is missing, invalid, or revoked             |
| 404         | `not_found`           | Endpoint does not exist                             |
| 405         | `method_not_allowed`  | HTTP method is not supported for this endpoint      |
| 429         | `rate_limit_exceeded` | Per-minute rate limit has been exceeded             |
| 500         | `internal_error`      | Unexpected server error                             |
| 502         | `geocoder_error`      | Upstream geocoder (Nominatim) returned an error     |
| 504         | `geocoder_timeout`    | Upstream geocoder did not respond in time           |

---

## Health and Readiness Endpoints

These endpoints do not require authentication.

### GET /healthz

Process liveness check. Returns `200 OK` when the API process is running.

```bash
curl http://localhost:8080/healthz
```

```json
{"status": "ok"}
```

### GET /readyz

Dependency readiness check. Returns `200 OK` only when Postgres, Redis, and Nominatim are all reachable.

```bash
curl http://localhost:8080/readyz
```

```json
HTTP/1.1 200 OK
{
  "status": "ok",
  "dependencies": {
    "postgres": "ok",
    "redis": "ok",
    "nominatim": "ok"
  }
}
```

When one or more dependencies are unhealthy:

```json
HTTP/1.1 503 Service Unavailable
{
  "status": "degraded",
  "dependencies": {
    "postgres": "ok",
    "redis": "ok",
    "nominatim": "error: connection refused"
  }
}
```

---

## Request IDs

Every response includes a `request_id` field (or `X-Request-ID` header). Use this value when reporting issues.

---

## Quick Start (Local Dev)

1. Start the local stack:
   ```bash
   docker compose -f deploy/compose/docker-compose.yml up -d
   ```

2. Create a tenant and API key using the seed script:
   ```bash
   go run scripts/seed.go
   ```
   The script prints the raw API key to stdout.

3. Test a forward geocode:
   ```bash
   curl -s -X POST http://localhost:8080/v1/geocode/forward \
     -H "Content-Type: application/json" \
     -H "X-API-Key: <your-key>" \
     -d '{"query": "Mairie de Monaco"}'
   ```
