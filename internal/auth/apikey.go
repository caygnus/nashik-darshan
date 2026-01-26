package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// HashAPIKey creates a SHA-256 hash of the API key
func HashAPIKey(key string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateAPIKey generates a new API key
// The key is returned in its raw form, it should be hashed before storing
func GenerateAPIKey() string {
	// Generate a random 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return hex.EncodeToString(key)
}
