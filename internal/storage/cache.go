package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache provides get/set operations for geocode response caching.
type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewCache creates a new Redis-backed cache.
func NewCache(addr string, ttl time.Duration) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Cache{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a cached value by key. Returns nil, nil on cache miss or error.
func (c *Cache) Get(ctx context.Context, key string) (json.RawMessage, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return json.RawMessage(val), nil
}

// Set stores a value in the cache with the configured TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// ForwardCacheKey generates a cache key for a forward geocode request.
func ForwardCacheKey(query string, limit int) string {
	raw := fmt.Sprintf("forward:%s:%d", query, limit)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("geocode:fwd:%x", hash[:16])
}

// ReverseCacheKey generates a cache key for a reverse geocode request.
// Coordinates are rounded to 5 decimal places (~1.1m precision).
func ReverseCacheKey(lat, lng float64) string {
	raw := fmt.Sprintf("reverse:%.5f:%.5f", lat, lng)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("geocode:rev:%x", hash[:16])
}

// Close closes the Redis connection.
func (c *Cache) Close() error {
	return c.client.Close()
}
