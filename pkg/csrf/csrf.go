package csrf

import (
	"crypto/rand"
	"encoding/hex"
)

// NewToken returns a random CSRF token (hex-encoded).
func NewToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
