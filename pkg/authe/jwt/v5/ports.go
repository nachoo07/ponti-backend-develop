package pkgjwt

import (
	"context"
	"time"
)

type Service interface {
	GenerateTokens(context.Context, string, time.Duration, time.Duration) (*Token, error)
	ValidateToken(context.Context, string) (*TokenClaims, error)
	GetAccessExpiration() time.Duration
	GetRefreshExpiration() time.Duration
	ValidateTokenAllowExpired(ctx context.Context, tokenString string) (*TokenClaims, error)
	ExtractClaimsFromExternalToken(tokenString string, signingMethod string, key any, claimKeys ...string) (map[string]any, error)
}

type Config interface {
	GetAccessExpiration() time.Duration
	GetRefreshExpiration() time.Duration
	GetSecretKey() string
	Validate() error
}
