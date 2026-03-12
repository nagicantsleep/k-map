package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter checks and enforces per-tenant rate limits.
type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// NewRateLimiter creates a Redis-backed rate limiter.
func NewRateLimiter(addr string, requestsPerMinute int) *RateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RateLimiter{
		client: client,
		limit:  requestsPerMinute,
		window: time.Minute,
	}
}

// Allow checks whether the tenant is within its rate limit.
// Returns true if the request is allowed, false if rate-limited.
func (rl *RateLimiter) Allow(ctx context.Context, tenantID string) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", tenantID)

	pipe := rl.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit check: %w", err)
	}

	count := incr.Val()

	return count <= int64(rl.limit), nil
}

// Close closes the Redis connection used by the rate limiter.
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}
