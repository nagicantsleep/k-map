package auth

import (
	"testing"
)

func TestHashKeyDeterministic(t *testing.T) {
	raw := "test-api-key-12345"
	h1 := HashKey(raw)
	h2 := HashKey(raw)

	if h1 != h2 {
		t.Fatalf("HashKey not deterministic: %s != %s", h1, h2)
	}

	if len(h1) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(h1))
	}
}

func TestHashKeyDifferentInputs(t *testing.T) {
	h1 := HashKey("key-a")
	h2 := HashKey("key-b")

	if h1 == h2 {
		t.Fatal("different inputs produced same hash")
	}
}

func TestGenerateRawKey(t *testing.T) {
	key, err := GenerateRawKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(key) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(key))
	}

	key2, err := GenerateRawKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key == key2 {
		t.Fatal("two generated keys should not be equal")
	}
}
