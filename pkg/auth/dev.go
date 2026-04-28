package auth

import (
	"context"
	"errors"
	"strings"
)

type DevVerifier struct{}

func (v DevVerifier) VerifyIDToken(_ context.Context, bearerToken string) (Claims, error) {
	// Dev mode accepts:
	// - "dev" (single shared identity)
	// - "dev:<uid>" or "dev:<uid>:<email>"
	if bearerToken == "" {
		return Claims{}, errors.New("missing bearer token")
	}
	if bearerToken == "dev" {
		return Claims{UID: "dev-user", Email: "dev@example.com"}, nil
	}
	if strings.HasPrefix(bearerToken, "dev:") {
		parts := strings.SplitN(strings.TrimPrefix(bearerToken, "dev:"), ":", 2)
		uid := strings.TrimSpace(parts[0])
		if uid == "" {
			return Claims{}, errors.New("invalid dev token")
		}
		claims := Claims{UID: uid}
		if len(parts) == 2 {
			claims.Email = strings.TrimSpace(parts[1])
		}
		return claims, nil
	}
	return Claims{}, errors.New("invalid dev token")
}
