#!/bin/bash
# import-osm.sh - Import OSM data into Nominatim
#
# Usage:
#   ./scripts/import-osm.sh
#   KMAP_NOMINATIM_PBF_URL=https://download.geofabrik.de/europe/belgium-latest.osm.pbf ./scripts/import-osm.sh
#
# Environment variables:
#   KMAP_NOMINATIM_PBF_URL  - URL of the OSM PBF file to import (default: Monaco)
#   KMAP_NOMINATIM_CONTAINER - Name of the Nominatim container (default: kmap-nominatim)
#   KMAP_COMPOSE_FILE        - Path to docker-compose.yml (default: deploy/compose/docker-compose.yml)

set -euo pipefail

# Configuration with defaults
PBF_URL="${KMAP_NOMINATIM_PBF_URL:-https://download.geofabrik.de/europe/monaco-latest.osm.pbf}"
CONTAINER="${KMAP_NOMINATIM_CONTAINER:-kmap-nominatim}"
COMPOSE_FILE="${KMAP_COMPOSE_FILE:-deploy/compose/docker-compose.yml}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

# Check if Docker Compose is available
if ! docker compose version &> /dev/null; then
    log_error "Docker Compose is not available"
    exit 1
fi

log_info "Starting OSM data import"
log_info "PBF URL: $PBF_URL"
log_info "Container: $CONTAINER"
log_info "Compose file: $COMPOSE_FILE"

# Stop the API service to prevent requests during import
log_info "Stopping API service..."
docker compose -f "$COMPOSE_FILE" stop api 2>/dev/null || true

# Stop and remove the existing Nominatim container
log_info "Stopping existing Nominatim container..."
docker compose -f "$COMPOSE_FILE" stop nominatim 2>/dev/null || true
docker compose -f "$COMPOSE_FILE" rm -f nominatim 2>/dev/null || true

# Remove Nominatim data volumes for a clean import
log_info "Removing existing Nominatim data volumes..."
docker volume rm kmap_nominatim-data 2>/dev/null || true
docker volume rm kmap_nominatim-flatnode 2>/dev/null || true

# Export the PBF URL for docker-compose
export KMAP_NOMINATIM_PBF_URL="$PBF_URL"

# Derive replication URL from PBF URL (for update support)
# Example: https://download.geofabrik.de/europe/monaco-latest.osm.pbf -> https://download.geofabrik.de/europe/monaco-updates/
REPLICATION_URL=$(echo "$PBF_URL" | sed 's/-latest\.osm\.pbf$/-updates\//')
export KMAP_NOMINATIM_REPLICATION_URL="$REPLICATION_URL"
log_info "Replication URL: $REPLICATION_URL"

# Start Nominatim with the new data
log_info "Starting Nominatim container with new data (this may take several minutes)..."
docker compose -f "$COMPOSE_FILE" up -d nominatim

# Wait for Nominatim to become healthy
log_info "Waiting for Nominatim to become healthy..."
TIMEOUT=1800  # 30 minutes
START_TIME=$(date +%s)

while true; do
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
    
    if [ $ELAPSED -ge $TIMEOUT ]; then
        log_error "Timeout waiting for Nominatim to become healthy"
        log_error "Check logs: docker logs $CONTAINER"
        exit 1
    fi
    
    # Check health status
    HEALTH=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER" 2>/dev/null || echo "unknown")
    
    if [ "$HEALTH" = "healthy" ]; then
        log_info "Nominatim is healthy after ${ELAPSED} seconds"
        break
    elif [ "$HEALTH" = "unhealthy" ]; then
        log_error "Nominatim health check failed"
        log_error "Check logs: docker logs $CONTAINER"
        exit 1
    fi
    
    echo -n "."
    sleep 10
done
echo

# Verify the import by checking the status endpoint
log_info "Verifying import..."
STATUS=$(curl -s "http://localhost:8081/status?format=json" 2>/dev/null || echo '{"status":-1}')

if echo "$STATUS" | grep -q '"status":0'; then
    log_info "Import verification successful"
else
    log_warn "Import verification returned non-zero status: $STATUS"
    log_warn "Check Nominatim logs: docker logs $CONTAINER"
fi

# Start the API service
log_info "Starting API service..."
docker compose -f "$COMPOSE_FILE" start api

log_info "OSM data import completed successfully"
log_info "Test with: curl -s 'http://localhost:8081/status?format=json' | jq"
