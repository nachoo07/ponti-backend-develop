package pkgjwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa las claims personalizadas para el token JWT.
type Claims struct {
	Subject string `json:"sub"`
	jwt.RegisteredClaims
}

// TokenClaims representa las claims extraídas de un token validado.
type TokenClaims struct {
	Subject   string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// Token define los datos que retornaremos al generar los tokens.
type Token struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	IssuedAt         time.Time
	Subject          string
	TokenType        string
}

