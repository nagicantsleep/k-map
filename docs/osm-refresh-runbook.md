# OSM Data Refresh Runbook

This runbook describes how to import OpenStreetMap data into the self-hosted Nominatim instance, verify the import, and roll back if necessary.

## Prerequisites

- Docker and Docker Compose installed
- Sufficient disk space (see sizing below)
- Network access to Geofabrik or alternative OSM extract provider

## Disk Sizing Guidelines

| Region | Approximate PBF Size | Recommended Disk (Nominatim) |
|--------|---------------------|------------------------------|
| Monaco | ~1 MB | 5 GB |
| Belgium | ~150 MB | 30 GB |
| Germany | ~3.5 GB | 150 GB |
| Europe | ~25 GB | 500 GB+ |

## Import Procedure

### 1. Choose an OSM Extract

Select a Geofabrik extract URL for your target region:

- Monaco (smallest, good for local testing): `https://download.geofabrik.de/europe/monaco-latest.osm.pbf`
- Full list: https://download.geofabrik.de/

### 2. Stop the API Service

Prevent requests to Nominatim during the import:

```bash
docker compose -f deploy/compose/docker-compose.yml stop api
```

### 3. Run the Import Script

```bash
# For Monaco (default)
./scripts/import-osm.sh

# For a different region, set the PBF URL
KMAP_NOMINATIM_PBF_URL=https://download.geofabrik.de/europe/belgium-latest.osm.pbf \
  ./scripts/import-osm.sh
```

The script will:
1. Download the PBF file if not cached
2. Delete existing Nominatim data
3. Start a fresh Nominatim container with the new data
4. Wait for the import to complete

### 4. Verify the Import

Check the Nominatim status endpoint:

```bash
curl -s http://localhost:8081/status?format=json | jq
```

Expected response:
```json
{
  "status": 0,
  "message": "OK"
}
```

If `status` is non-zero, see the Nominatim logs:

```bash
docker logs kmap-nominatim
```

### 5. Restart the API Service

```bash
docker compose -f deploy/compose/docker-compose.yml start api
```

### 6. Functional Verification

Test forward geocoding:

```bash
curl -X POST http://localhost:8080/v1/geocode/forward \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-key" \
  -d '{"query":"Monaco"}'
```

## Rollback Procedure

If the import fails or produces incorrect results:

### 1. Stop Services

```bash
docker compose -f deploy/compose/docker-compose.yml down
```

### 2. Remove Nominatim Data Volumes

```bash
docker volume rm kmap_nominatim-data kmap_nominatim-flatnode
```

### 3. Restore from Backup (if available)

If you have a backup of the Nominatim data volume:

```bash
# Example: restore from a tar archive
docker run --rm -v kmap_nominatim-data:/data -v /path/to/backup:/backup alpine \
  tar xzf /backup/nominatim-data.tar.gz -C /data
```

### 4. Recreate Stack

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## Update Cadence

For production deployments, consider:

- **Weekly updates**: Small regions (Monaco, Luxembourg)
- **Monthly updates**: Medium regions (Belgium, Netherlands)
- **Quarterly updates**: Large regions (Germany, France, full Europe)

Update frequency depends on:
- Rate of OSM changes in the region
- Business requirements for data freshness
- Available maintenance windows

## Automation

For scheduled updates, consider:

1. A cron job that runs the import script during low-traffic periods
2. Monitoring that alerts on import failures
3. Automated rollback if post-import health checks fail

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KMAP_NOMINATIM_PBF_URL` | URL of the OSM PBF file to import | Monaco |
| `KMAP_NOMINATIM_REPLICATION_URL` | URL for Nominatim replication updates | Monaco updates |
| `KMAP_NOMINATIM_CONTAINER` | Name of the Nominatim container | `kmap-nominatim` |

## Troubleshooting

### Import hangs at "Importing..."

Check available memory and disk space:

```bash
docker stats kmap-nominatim
df -h
```

Nominatim is memory-intensive. For large regions, ensure:
- At least 2 GB RAM per import thread
- SSD storage for the data volume

### "Connection refused" after import

The Nominatim container may still be initializing. Wait for the health check to pass:

```bash
docker compose -f deploy/compose/docker-compose.yml logs -f nominatim
```

### Search returns no results

1. Verify the PBF file matches your expected region
2. Check that the data import completed without errors
3. Test a known address in the imported region

## References

- [Nominatim Documentation](https://nominatim.org/release-docs/latest/)
- [Geofabrik Downloads](https://download.geofabrik.de/)
- [OSM Wiki: Nominatim](https://wiki.openstreetmap.org/wiki/Nominatim)
