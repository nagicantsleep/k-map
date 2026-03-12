package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const rawKeyLength = 32

// GenerateRawKey produces a cryptographically random API key string.
func GenerateRawKey() (string, error) {
	b := make([]byte, rawKeyLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// HashKey returns the SHA-256 hex digest of a raw API key.
func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
