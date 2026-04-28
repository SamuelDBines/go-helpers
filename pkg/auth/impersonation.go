package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const impersonationIssuer = "relayemail/impersonation"

type impersonationClaims struct {
	jwt.RegisteredClaims
	UID           string `json:"uid"`
	Email         string `json:"email,omitempty"`
	ImpersonatedBy string `json:"impersonatedBy,omitempty"`
}

// CreateImpersonationToken creates a short-lived JWT for viewing as another user.
func CreateImpersonationToken(secret string, uid, email, adminUID string, exp time.Duration) (string, error) {
	if len(secret) < 32 {
		return "", errors.New("impersonation secret must be at least 32 characters")
	}
	now := time.Now().UTC()
	claims := impersonationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    impersonationIssuer,
			Subject:   uid,
			ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UID:           uid,
		Email:         email,
		ImpersonatedBy: adminUID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ImpersonationVerifier verifies impersonation JWTs and returns the impersonated user's claims.
type ImpersonationVerifier struct {
	secret []byte
}

// NewImpersonationVerifier creates a verifier for impersonation tokens.
func NewImpersonationVerifier(secret string) (*ImpersonationVerifier, error) {
	if len(secret) < 32 {
		return nil, errors.New("impersonation secret must be at least 32 characters")
	}
	return &ImpersonationVerifier{secret: []byte(secret)}, nil
}

// VerifyIDToken verifies an impersonation JWT and returns the impersonated user's claims.
func (v *ImpersonationVerifier) VerifyIDToken(_ context.Context, bearerToken string) (Claims, error) {
	if bearerToken == "" {
		return Claims{}, errors.New("missing bearer token")
	}
	parsed, err := jwt.ParseWithClaims(bearerToken, &impersonationClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}
	c, ok := parsed.Claims.(*impersonationClaims)
	if !ok || c.Issuer != impersonationIssuer {
		return Claims{}, errors.New("invalid impersonation claims")
	}
	return Claims{UID: c.UID, Email: c.Email}, nil
}
