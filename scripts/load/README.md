# Load Testing

This directory contains load test scripts for the k-map API.

## Prerequisites

Install [k6](https://k6.io/docs/get-started/installation/):

```bash
# macOS
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" \
  | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update && sudo apt-get install k6

# Windows (Chocolatey)
choco install k6

# Docker
docker run --rm -i grafana/k6 run - <scripts/load/k6_load_test.js
```

## Running the Load Test

Ensure the local stack is running first:

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

### Against the local stack (default)

```bash
k6 run scripts/load/k6_load_test.js
```

### Against a deployed API

```bash
KMAP_API_URL=https://your-api.example.com KMAP_API_KEY=your-key k6 run scripts/load/k6_load_test.js
```

### Via Docker (no local k6 install needed)

```bash
docker run --rm --network host \
  -e KMAP_API_URL=http://localhost:8080 \
  -e KMAP_API_KEY=your-key \
  -v $(pwd)/scripts/load:/scripts \
  grafana/k6 run /scripts/k6_load_test.js
```

## Test Parameters

| Parameter | Value |
|-----------|-------|
| Concurrent VUs | 50 |
| Duration | 2 minutes |
| Dataset | Monaco (Geofabrik extract) |
| Endpoints | Forward, Reverse, Proximity (round-robin) |

## PRD Latency Targets (Pass/Fail Thresholds)

| Endpoint | P95 Target |
|----------|-----------|
| Forward geocode | < 800 ms |
| Reverse geocode | < 500 ms |
| Proximity check | < 1000 ms |
| 5xx error rate | < 1% |

k6 will exit with a non-zero code if any threshold is violated.

## Interpreting Results

A passing run produces output like:

```
✓ forward: status 200
✓ reverse: status 200
✓ proximity: status 200

forward_duration......: avg=120ms p(90)=310ms p(95)=450ms
reverse_duration......: avg=80ms  p(90)=180ms p(95)=240ms
proximity_duration....: avg=150ms p(90)=380ms p(95)=550ms

✓ forward_duration{p(95)}: p(95)<800
✓ reverse_duration{p(95)}: p(95)<500
✓ proximity_duration{p(95)}: p(95)<1000
✓ http_req_failed: rate<0.01
```

Post the k6 summary output as a comment on issue #44 when running as part of the beta readiness checklist.

## Adding New Load Scenarios

Edit `k6_load_test.js` and add a new `scenarios` entry in `options.scenarios`. Common patterns:

- **Ramp-up:** Gradually increase VUs to find breaking point
- **Spike test:** Sudden burst of traffic
- **Soak test:** Sustained low load over hours to detect memory leaks
