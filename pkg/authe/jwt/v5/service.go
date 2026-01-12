package pkgjwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// service implementa la interfaz Service con un único secret y posibles expiraciones personalizadas.
type service struct {
	config            Config
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

// newService crea e inicializa un nuevo Service a partir de la configuración (un solo secret).
func newService(c Config) (Service, error) {
	return &service{
		config:            c,
		secret:            []byte(c.GetSecretKey()),
		accessExpiration:  c.GetAccessExpiration(),
		refreshExpiration: c.GetRefreshExpiration(),
	}, nil
}

// GenerateTokens crea un par de tokens (access y refresh) usando un único secret.
// Se permiten expiraciones custom (customAccessExp, customRefreshExp) que, si no son 0,
// sobreescriben las expiraciones por defecto definidas en la configuración.
func (s *service) GenerateTokens(ctx context.Context, subject string,
	customAccessExp, customRefreshExp time.Duration) (*Token, error) {

	now := time.Now()

	// Calcular expiraciones (o usar las de la config)
	accessExp := s.accessExpiration
	if customAccessExp != 0 {
		accessExp = customAccessExp
	}
	refreshExp := s.refreshExpiration
	if customRefreshExp != 0 {
		refreshExp = customRefreshExp
	}

	accessTokenExpiresAt := now.Add(accessExp)
	refreshTokenExpiresAt := now.Add(refreshExp)

	// Generar el access token
	accessClaims := Claims{
		Subject: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("error signing the access token: %w", err)
	}

	// Generar el refresh token
	refreshClaims := Claims{
		Subject: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("error signing the refresh token: %w", err)
	}

	// Retornar los tokens generados

	return &Token{
		AccessToken:      signedAccessToken,
		RefreshToken:     signedRefreshToken,
		AccessExpiresAt:  accessTokenExpiresAt,
		RefreshExpiresAt: refreshTokenExpiresAt,
		IssuedAt:         now,
		Subject:          subject,
		TokenType:        "Bearer",
	}, nil
}

// ValidateToken valida un token y retorna las claims extraídas, usando el único secret.
func (s *service) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Verificamos que sea un método de firma esperado (HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error validating the token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	tokenClaims := &TokenClaims{
		Subject:   claims.Subject,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}
	return tokenClaims, nil
}

// ValidateTokenAllowExpired valida el token pero permite que esté expirado.
// Retorna las claims si el token estaba correctamente firmado y parseado,
// incluso si ocurrió el error de expiración.
func (s *service) ValidateTokenAllowExpired(ctx context.Context, tokenString string) (*TokenClaims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		// Verificamos si el error se debe únicamente a expiración
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Devolvemos las claims aunque el token esté expirado
			return &TokenClaims{
				Subject:   claims.Subject,
				ExpiresAt: claims.ExpiresAt.Time,
				IssuedAt:  claims.IssuedAt.Time,
			}, nil
		}
		return nil, fmt.Errorf("error validating the token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &TokenClaims{
		Subject:   claims.Subject,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}, nil
}

// GetAccessExpiration expone la expiración del access token desde la configuración.
func (s *service) GetAccessExpiration() time.Duration {
	return s.config.GetAccessExpiration()
}

// GetRefreshExpiration expone la expiración del refresh token desde la configuración.
func (s *service) GetRefreshExpiration() time.Duration {
	return s.config.GetRefreshExpiration()
}

func (s *service) ExtractClaimsFromExternalToken(tokenString string, signingMethod string, key any, claimKeys ...string) (map[string]any, error) {
	var keyFunc jwt.Keyfunc

	switch strings.ToUpper(signingMethod) {
	case "HMAC":
		secretKey, ok := key.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid secret key for HMAC")
		}
		keyFunc = func(token *jwt.Token) (any, error) {
			// Verificar el método de firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		}
	case "RSA":
		publicKeyPEM, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("invalid PEM public key for RSA")
		}
		// Decodificar el bloque PEM
		block, _ := pem.Decode([]byte(publicKeyPEM))
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("failed to decode PEM block")
		}

		// Parsear la clave pública
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}

		rsaPublicKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not a valid RSA public key")
		}

		keyFunc = func(token *jwt.Token) (any, error) {
			// Verificar el método de firma
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return rsaPublicKey, nil
		}
	default:
		return nil, fmt.Errorf("unknown signing method: %s", signingMethod)
	}

	// Parsear el token
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extraer los claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Opcional: Validar claims estándar como 'exp' (expiración)
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return nil, fmt.Errorf("token has expired")
		}
	}

	// Extraer las claims especificadas
	extractedClaims := make(map[string]any)
	for _, claimKey := range claimKeys {
		if claimValue, exists := claims[claimKey]; exists {
			extractedClaims[claimKey] = claimValue
		} else {
			return nil, fmt.Errorf("claim '%s' not found in token", claimKey)
		}
	}

	return extractedClaims, nil
}
