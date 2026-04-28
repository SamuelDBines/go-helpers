package auth

import "context"

// ChainVerifier tries verifiers in order; returns the first successful result.
type ChainVerifier []Verifier

// VerifyIDToken tries each verifier in order.
func (c ChainVerifier) VerifyIDToken(ctx context.Context, bearerToken string) (Claims, error) {
	var lastErr error
	for _, v := range c {
		if v == nil {
			continue
		}
		claims, err := v.VerifyIDToken(ctx, bearerToken)
		if err == nil {
			return claims, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return Claims{}, lastErr
	}
	return Claims{}, nil
}
