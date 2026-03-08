# Local Compose Stack

This compose stack provides the first local runtime slice for `k-map`:

- `api`: the Go HTTP service from `cmd/api`
- `postgres`: application metadata database
- `redis`: cache and rate-limit store
- `nominatim`: self-hosted geocoder runtime for forward and reverse lookups

## Commands

Start the stack:

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

Stop the stack:

```bash
docker compose -f deploy/compose/docker-compose.yml down
```

Validate the compose configuration without starting containers:

```bash
docker compose -f deploy/compose/docker-compose.yml config
```

## Local Runtime Defaults

The API container uses explicit dependency configuration so the same binary can run either on the
host against published ports or inside the compose network:

- `KMAP_HTTP_ADDR=:8080`
- `KMAP_POSTGRES_ADDR=localhost:5432` by default, overridden to `postgres:5432` in compose
- `KMAP_REDIS_ADDR=localhost:6379` by default, overridden to `redis:6379` in compose
- `KMAP_NOMINATIM_URL=http://localhost:8081` by default, overridden to `http://nominatim:8080` in compose

The stateful services expose stable local service names on the `kmap` network for later wiring:

- `postgres:5432`
- `redis:6379`
- `nominatim:8080`

Postgres uses the baseline local development credentials:

- database: `kmap`
- user: `kmap`
- password: `kmap`

Persistent storage is currently enabled for Postgres through the named volume `postgres-data`.
Redis is configured as an in-memory development cache with persistence disabled.
Nominatim persists its embedded Postgres cluster and flatnode file through the named volumes
`nominatim-data` and `nominatim-flatnode`.

## Nominatim Import Defaults

The local stack now boots `mediagis/nominatim:5.2` with a lightweight Monaco extract by default so
the first import remains practical in local development.

Default dataset:

- `https://download.geofabrik.de/europe/monaco-latest.osm.pbf`

Default replication source:

- `https://download.geofabrik.de/europe/monaco-updates/`

These defaults can be overridden per shell session before running compose:

```bash
export KMAP_NOMINATIM_PBF_URL=https://download.geofabrik.de/north-america/us/california-latest.osm.pbf
export KMAP_NOMINATIM_REPLICATION_URL=https://download.geofabrik.de/north-america/us/california-updates/
docker compose -f deploy/compose/docker-compose.yml up -d
```

On PowerShell, set the same values with:

```powershell
$env:KMAP_NOMINATIM_PBF_URL = "https://download.geofabrik.de/north-america/us/california-latest.osm.pbf"
$env:KMAP_NOMINATIM_REPLICATION_URL = "https://download.geofabrik.de/north-america/us/california-updates/"
docker compose -f deploy/compose/docker-compose.yml up -d
```

Import notes:

- The first `nominatim` startup downloads the configured `.osm.pbf` extract and performs the import inside the container.
- Follow progress with `docker logs -f kmap-nominatim`.
- Query the local geocoder on `http://localhost:8081` after the import completes.
- Keep `FREEZE=true` for the current baseline so local environments do not start replication unexpectedly.

## Storage Expectations

- Monaco is appropriate for local bring-up and smoke tests because it keeps import time and disk usage low.
- Larger Geofabrik extracts will increase startup time, disk, and memory needs materially; treat them as opt-in overrides rather than the default developer path.
- `/readyz` now verifies TCP reachability for Postgres, Redis, and Nominatim based on the configured dependency endpoints.
