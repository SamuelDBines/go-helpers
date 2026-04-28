package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

type FirebaseVerifier struct {
	projectID string
	issuer    string

	mu   sync.Mutex
	jwks *keyfunc.JWKS
}

type FirebaseVerifierOptions struct {
	ProjectID string
}

func NewFirebaseVerifier(_ context.Context, opt FirebaseVerifierOptions) (*FirebaseVerifier, error) {
	if opt.ProjectID == "" {
		return nil, errors.New("ProjectID required")
	}
	return &FirebaseVerifier{
		projectID: opt.ProjectID,
		issuer:    fmt.Sprintf("https://securetoken.google.com/%s", opt.ProjectID),
	}, nil
}

func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, bearerToken string) (Claims, error) {
	if bearerToken == "" {
		return Claims{}, errors.New("missing bearer token")
	}

	jwks, err := v.getJWKS(ctx)
	if err != nil {
		return Claims{}, err
	}

	parsed, err := jwt.Parse(bearerToken, jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("invalid claims")
	}

	if err := validateFirebaseClaims(claims, v.projectID, v.issuer); err != nil {
		return Claims{}, err
	}

	uid, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if uid == "" {
		return Claims{}, errors.New("missing sub")
	}

	return Claims{UID: uid, Email: email}, nil
}

func (v *FirebaseVerifier) getJWKS(ctx context.Context) (*keyfunc.JWKS, error) {
	// Firebase (Secure Token Service) JWKS.
	// Docs commonly reference the x509 endpoint; this JWKS endpoint works for RS256 verification.
	const jwksURL = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.jwks != nil {
		return v.jwks, nil
	}

	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  5 * time.Minute,
		RefreshTimeout:    10 * time.Second,
		RefreshUnknownKID: true,
		Ctx:               ctx,
	})
	if err != nil {
		return nil, err
	}
	v.jwks = jwks
	return jwks, nil
}

func validateFirebaseClaims(c jwt.MapClaims, projectID, issuer string) error {
	now := time.Now().Unix()

	iss, _ := c["iss"].(string)
	if iss != issuer {
		return errors.New("invalid iss")
	}

	aud, ok := c["aud"].(string)
	if !ok || aud != projectID {
		return errors.New("invalid aud")
	}

	if exp, ok := c["exp"].(float64); ok {
		if int64(exp) <= now {
			return errors.New("token expired")
		}
	} else {
		return errors.New("missing exp")
	}

	if iat, ok := c["iat"].(float64); ok {
		if int64(iat) > now+60 {
			return errors.New("iat in the future")
		}
	} else {
		return errors.New("missing iat")
	}

	sub, _ := c["sub"].(string)
	if sub == "" {
		return errors.New("missing sub")
	}

	return nil
}
