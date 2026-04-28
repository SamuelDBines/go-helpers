package auth

import "context"

type Claims struct {
	UID   string
	Email string
}

type Verifier interface {
	VerifyIDToken(ctx context.Context, bearerToken string) (Claims, error)
}
