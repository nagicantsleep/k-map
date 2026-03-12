package storage

import (
	"testing"
)

func TestForwardCacheKey_Deterministic(t *testing.T) {
	t.Parallel()

	key1 := ForwardCacheKey("test query", 5)
	key2 := ForwardCacheKey("test query", 5)

	if key1 != key2 {
		t.Errorf("expected same key, got %s and %s", key1, key2)
	}
}

func TestForwardCacheKey_DifferentInputs(t *testing.T) {
	t.Parallel()

	key1 := ForwardCacheKey("query a", 5)
	key2 := ForwardCacheKey("query b", 5)

	if key1 == key2 {
		t.Error("expected different keys for different queries")
	}
}

func TestForwardCacheKey_DifferentLimits(t *testing.T) {
	t.Parallel()

	key1 := ForwardCacheKey("test", 5)
	key2 := ForwardCacheKey("test", 10)

	if key1 == key2 {
		t.Error("expected different keys for different limits")
	}
}

func TestReverseCacheKey_Deterministic(t *testing.T) {
	t.Parallel()

	key1 := ReverseCacheKey(37.42200, -122.08400)
	key2 := ReverseCacheKey(37.42200, -122.08400)

	if key1 != key2 {
		t.Errorf("expected same key, got %s and %s", key1, key2)
	}
}

func TestReverseCacheKey_RoundsCoordinates(t *testing.T) {
	t.Parallel()

	// These should round to the same 5-decimal value
	key1 := ReverseCacheKey(37.422001, -122.084001)
	key2 := ReverseCacheKey(37.422002, -122.084002)

	// At 5 decimal places these round differently
	// 37.42200 vs 37.42200 — actually same at 5dp
	if key1 != key2 {
		t.Errorf("expected same key for coordinates within rounding, got %s and %s", key1, key2)
	}
}

func TestReverseCacheKey_DifferentCoordinates(t *testing.T) {
	t.Parallel()

	key1 := ReverseCacheKey(37.422, -122.084)
	key2 := ReverseCacheKey(48.856, 2.352)

	if key1 == key2 {
		t.Error("expected different keys for different coordinates")
	}
}
