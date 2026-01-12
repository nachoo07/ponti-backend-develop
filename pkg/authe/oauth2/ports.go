package pkgoauth2

import (
	"context"
	"time"
)

// Config define la interfaz para configuración básica de OAuth2.
type Config interface {
	Validate() error
	GetClientID() string
	GetClientSecret() string
	GetAuthURL() string
	GetTokenURL() string
	GetRedirectURL() string
	GetScopes() []string
	GetTimeout() time.Duration
}

// Service define la interfaz para un servicio OAuth2 genérico.
type Service interface {
	// Construye la URL para obtener el código de autorización (Authorization Code Flow).
	GetAuthCodeURL(state string) string

	// Intercambia un código de autorización por un token.
	ExchangeCode(ctx context.Context, code string) (*OAuth2Token, error)

	// Usa el refresh token para obtener un nuevo access token.
	RefreshToken(ctx context.Context, refreshToken string) (*OAuth2Token, error)

	// Valida un token (depende del proveedor).
	ValidateToken(ctx context.Context, tokenStr string) (*TokenClaims, error)
}
