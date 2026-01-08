package jwtmiddleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

// Validate retorna un http.HandlerFunc que valida el JWT usando la lógica común y llama a 'next' si la validación es exitosa.
func Validate(cfg pkgutils.Config, next http.HandlerFunc) http.HandlerFunc {
	// Parsear la clave RSA si se proporciona.
	var rsaPublicKey *rsa.PublicKey
	if cfg.PublicKeyPEM != "" {
		key, err := pkgutils.ParseRSAPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			log.Fatalf("failed to parse RSA public key: %v", err)
		}
		rsaPublicKey = key
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := pkgutils.ExtractTokenFromRequest(r, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Parsear el token sin verificar para conocer el método de firma.
		unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		keyFunc := pkgutils.SelectKeyFunc(unverifiedToken, cfg.SecretKey, rsaPublicKey)
		if keyFunc == nil {
			http.Error(w, "unexpected signing method", http.StatusUnauthorized)
			return
		}

		parsedToken, err := jwt.Parse(tokenStr, keyFunc)
		if err != nil || !parsedToken.Valid {
			http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Añadir el token y los claims al contexto de la request.
		ctx := r.Context()
		type contextKey string
		ctx = context.WithValue(ctx, contextKey(cfg.ContextKey), parsedToken)
		ctx = context.WithValue(ctx, contextKey(pkgutils.GetClaimsKey(cfg.ContextKey)), parsedToken.Claims)
		next(w, r.WithContext(ctx))
	}
}
