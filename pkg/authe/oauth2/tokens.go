package pkgoauth2

import (
	"time"
)

// OAuth2Token guarda información relevante del token OAuth2
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

// TokenClaims representa claims opcionales (p.ej. JWT)
type TokenClaims struct {
	Subject string `json:"sub,omitempty"`
	// Otros campos como email, nombre, roles, etc.
}
