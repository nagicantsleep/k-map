# Local Compose Stack

This compose stack provides the first local runtime slice for `k-map`:

- `api`: the Go HTTP service from `cmd/api`
- `postgres`: application metadata database
- `redis`: cache and rate-limit store

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

The API container currently only needs the HTTP bind address for this slice:

- `KMAP_HTTP_ADDR=:8080`

The stateful services expose stable local service names on the `kmap` network for later wiring:

- `postgres:5432`
- `redis:6379`

Postgres uses the baseline local development credentials:

- database: `kmap`
- user: `kmap`
- password: `kmap`

Persistent storage is currently enabled for Postgres through the named volume `postgres-data`.
Redis is configured as an in-memory development cache with persistence disabled.

## Notes

- This issue intentionally does not add `Nominatim` yet. That arrives in the next epic sub-issue so the compose topology can be extended without mixing concerns.
- The API service does not yet consume Postgres or Redis connection settings. Those will be introduced when dependency wiring is added.
