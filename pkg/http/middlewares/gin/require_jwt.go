package pkgmwr

import (
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	pkgutils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

// RequireJWT valida el token JWT en la solicitud.
func RequireJWT(cfg pkgutils.Config) gin.HandlerFunc {
	var rsaPublicKey *rsa.PublicKey
	if cfg.PublicKeyPEM != "" {
		key, err := pkgutils.ParseRSAPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			panic(fmt.Errorf("failed to parse RSA public key: %w", err))
		}
		rsaPublicKey = key
	}

	return func(c *gin.Context) {
		tokenStr, err := pkgutils.ExtractTokenFromRequest(c.Request, cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "INVALID_TOKEN"})
			return
		}
		keyFunc := pkgutils.SelectKeyFunc(unverifiedToken, cfg.SecretKey, rsaPublicKey)
		if keyFunc == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNEXPECTED_SIGNING_METHOD"})
			return
		}
		parsedToken, err := jwt.Parse(tokenStr, keyFunc)
		if err != nil || !parsedToken.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "INVALID_TOKEN"})
			return
		}
		c.Set(cfg.ContextKey, parsedToken)
		c.Set(pkgutils.GetClaimsKey(cfg.ContextKey), parsedToken.Claims)
		c.Next()
	}
}
