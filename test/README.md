# Test Suite

This directory contains integration tests and golden dataset validation for the k-map API.

## Test Files

| File | Tag | Description |
|------|-----|-------------|
| `geocode_integration_test.go` | `integration` | Basic integration tests for forward and reverse geocoding against a live Nominatim stack |
| `auth_integration_test.go` | *(unit)* | Auth middleware integration tests using in-memory stubs |
| `golden_dataset_test.go` | `integration` | Golden dataset validation for the Monaco launch region |

## Golden Dataset Validation

The golden dataset tests validate all three API endpoints (forward geocoding, reverse geocoding, proximity validation) against a fixture file of known-good cases for the Monaco Geofabrik extract.

### Fixture File

`test/fixtures/monaco_golden.json` contains:
- **5 forward geocoding cases** — queries with expected coordinate bounding boxes
- **5 reverse geocoding cases** — known coordinates expected to resolve to an address
- **5 proximity cases** — labeled near/not-near pairs with explicit thresholds

### Running the Golden Dataset Tests

Ensure the local stack is running first:

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

Then run with the `integration` build tag:

```bash
go test -v -tags integration ./test/... -run TestGoldenDataset
```

To run against a deployed API instead of the local httptest server, set the base URL:

```bash
KMAP_API_URL=http://my-deployed-api:8080 go test -v -tags integration ./test/... -run TestGoldenDataset
```

### Adding New Fixture Cases

Edit `test/fixtures/monaco_golden.json` and add entries to the appropriate array:

**Forward case:**
```json
{
  "id": "fw-06",
  "query": "Some Monaco Place",
  "expect_lat_min": 43.70,
  "expect_lat_max": 43.75,
  "expect_lon_min": 7.40,
  "expect_lon_max": 7.45,
  "expect_min_results": 1
}
```

**Reverse case:**
```json
{
  "id": "rv-06",
  "description": "Some Monaco location",
  "latitude": 43.7300,
  "longitude": 7.4200,
  "expect_result": true
}
```

**Proximity case:**
```json
{
  "id": "px-06",
  "description": "Near/not-near scenario description",
  "latitude": 43.7311,
  "longitude": 7.4197,
  "target_query": "Target address or place",
  "threshold_meters": 200,
  "expect_is_near": true
}
```

### Adding Cases for a New Geography

1. Create a new fixture file under `test/fixtures/`, e.g. `test/fixtures/paris_golden.json` using the same JSON schema.
2. Add a corresponding `_test.go` file or extend `golden_dataset_test.go` to load additional fixture files via an environment variable or flag.

## Running All Integration Tests

```bash
go test -v -tags integration ./test/...
```

## Running Unit Tests Only

```bash
go test ./...
```
