package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashToken deterministically hashes arbitrary-length secrets (e.g., refresh tokens)
// to safely store them in the database without leaking their raw value.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CompareTokenHash checks if the provided token matches the stored hash using
// constant-time comparison to avoid timing attacks.
func CompareTokenHash(hash, token string) bool {
	if hash == "" || token == "" {
		return false
	}
	candidate := HashToken(token)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(candidate)) == 1
}
