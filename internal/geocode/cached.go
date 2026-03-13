package geocode

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nagicantsleep/k-map/internal/api"
	"github.com/nagicantsleep/k-map/internal/storage"
	"github.com/nagicantsleep/k-map/internal/telemetry"
)

// CachedGeocoder wraps a Geocoder with Redis caching.
// Cache failures degrade gracefully to direct geocoder calls.
type CachedGeocoder struct {
	inner   api.Geocoder
	cache   *storage.Cache
	logger  *slog.Logger
	metrics *telemetry.Metrics
}

// NewCachedGeocoder creates a caching wrapper around a Geocoder.
func NewCachedGeocoder(inner api.Geocoder, cache *storage.Cache, logger *slog.Logger) *CachedGeocoder {
	return &CachedGeocoder{
		inner:  inner,
		cache:  cache,
		logger: logger,
	}
}

// WithMetrics attaches a metrics collector to the CachedGeocoder.
func (g *CachedGeocoder) WithMetrics(m *telemetry.Metrics) *CachedGeocoder {
	g.metrics = m
	return g
}

// Search performs a forward geocoding search with caching.
func (g *CachedGeocoder) Search(ctx context.Context, query string, limit int) ([]api.GeocodeResult, error) {
	key := storage.ForwardCacheKey(query, limit)

	cached, err := g.cache.Get(ctx, key)
	if err != nil {
		g.logger.Warn("cache get failed", "key", key, "error", err)
	}

	if cached != nil {
		var results []api.GeocodeResult
		if err := json.Unmarshal(cached, &results); err == nil {
			if g.metrics != nil {
				g.metrics.CacheHits.WithLabelValues("search").Inc()
			}
			return results, nil
		}
		g.logger.Warn("cache unmarshal failed", "key", key, "error", err)
	}

	if g.metrics != nil {
		g.metrics.CacheMisses.WithLabelValues("search").Inc()
	}

	results, err := g.inner.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	if err := g.cache.Set(ctx, key, results); err != nil {
		g.logger.Warn("cache set failed", "key", key, "error", err)
	}

	return results, nil
}

// Reverse performs a reverse geocoding lookup with caching.
func (g *CachedGeocoder) Reverse(ctx context.Context, lat, lng float64) (*api.GeocodeResult, error) {
	key := storage.ReverseCacheKey(lat, lng)

	cached, err := g.cache.Get(ctx, key)
	if err != nil {
		g.logger.Warn("cache get failed", "key", key, "error", err)
	}

	if cached != nil {
		var result api.GeocodeResult
		if err := json.Unmarshal(cached, &result); err == nil {
			if g.metrics != nil {
				g.metrics.CacheHits.WithLabelValues("reverse").Inc()
			}
			return &result, nil
		}
		g.logger.Warn("cache unmarshal failed", "key", key, "error", err)
	}

	if g.metrics != nil {
		g.metrics.CacheMisses.WithLabelValues("reverse").Inc()
	}

	result, err := g.inner.Reverse(ctx, lat, lng)
	if err != nil {
		return nil, err
	}

	if result != nil {
		if err := g.cache.Set(ctx, key, result); err != nil {
			g.logger.Warn("cache set failed", "key", key, "error", err)
		}
	}

	return result, nil
}
