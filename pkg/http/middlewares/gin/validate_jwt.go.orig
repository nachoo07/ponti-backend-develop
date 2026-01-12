package pkgmwr

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

/* ---------------------------------------------------------------------------
   1) Constantes, mensajes de error y valores por defecto
   --------------------------------------------------------------------------- */

// Claves de contexto por defecto, prefijos y mensajes de error
const (
	DefaultContextKey          = "token"         // Clave por defecto para guardar el token en el contexto
	ClaimsSuffix               = "_claims"       // Sufijo para guardar los claims
	authHeaderName             = "Authorization" // Nombre del header para el token
	bearerPrefix               = "Bearer "       // Prefijo habitual en Authorization
	errMissingAuthHeader       = "authorization header required"
	errInvalidSigningMethod    = "unexpected signing method"
	errBearerPrefixRequired    = "authorization header must start with Bearer"
	errInvalidToken            = "invalid token"
	errExpiredToken            = "token has expired"
	errTokenNotFound           = "token not found in context"
	errClaimNotFound           = "claim not found in token"
	errInvalidClaimType        = "invalid claim type"
	errInvalidTokenLookup      = "invalid token lookup config"
	errUnsupportedLookupMethod = "unsupported token lookup method"
)

/* ---------------------------------------------------------------------------
   2) Estructura de configuración (Config) y constructor opcional
   --------------------------------------------------------------------------- */

// Config define cómo el middleware JWT se comportará:
//   - SecretKey para tokens HMAC
//   - PublicKeyPEM para tokens RSA
//   - TokenLookup determina de dónde se extrae el token (header/query)
//   - TokenPrefix (ej. "Bearer ")
//   - ContextKey para guardar el token parseado en gin.Context
type Config struct {
	SecretKey    string
	PublicKeyPEM string
	TokenLookup  string
	TokenPrefix  string
	ContextKey   string
}

// NewConfigFromEnv crea una configuración leyendo variables de entorno,
// asignando algunos defaults cuando no existan. Es opcional, puedes crear
// tu Config manualmente si prefieres.
func NewConfigFromEnv() Config {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	publicKey := os.Getenv("JWT_PUBLIC_PEM_KEY")
	contextKey := os.Getenv("JWT_CONTEXT_KEY")

	// Asignar valores por defecto
	if contextKey == "" {
		contextKey = DefaultContextKey
	}

	return Config{
		SecretKey:    secretKey,
		PublicKeyPEM: publicKey,
		// Por defecto, buscaremos en el header Authorization
		TokenLookup: "header:" + authHeaderName,
		TokenPrefix: bearerPrefix,
		ContextKey:  contextKey,
	}
}

/* ---------------------------------------------------------------------------
   3) Middleware principal: Validate
   --------------------------------------------------------------------------- */

// Validate recibe un Config y retorna un gin.HandlerFunc. Sirve como middleware
// de validación JWT para las rutas que lo requieran.
func Validate(cfg Config) gin.HandlerFunc {
	// Configuración por defecto
	if cfg.TokenLookup == "" {
		cfg.TokenLookup = "header:" + authHeaderName
	}
	if cfg.TokenPrefix == "" {
		cfg.TokenPrefix = bearerPrefix
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = DefaultContextKey
	}

	// Parsear la clave pública RSA
	var rsaPublicKey *rsa.PublicKey
	if cfg.PublicKeyPEM != "" {
		key, err := parseRSAPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			// Mejor manejo del error
			log.Fatalf("failed to parse RSA public key: %v", err)
		}
		rsaPublicKey = key
	}

	return func(c *gin.Context) {
		log.Println("JWT Middleware: Starting validation...")

		tokenStr, err := extractToken(c, cfg)
		if err != nil {
			abortWithError(c, http.StatusUnauthorized, err.Error())
			return
		}

		// Parsear sin verificar para determinar el método de firma
		unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			abortWithError(c, http.StatusUnauthorized, fmt.Sprintf("%s: %v", errInvalidToken, err))
			return
		}

		// Seleccionar la función de clave adecuada
		keyFunc := selectKeyFunc(unverifiedToken, cfg.SecretKey, rsaPublicKey)
		if keyFunc == nil {
			abortWithError(c, http.StatusUnauthorized, errInvalidSigningMethod)
			return
		}

		// Parsear y validar el token
		parsedToken, err := jwt.Parse(tokenStr, keyFunc)
		if err != nil || !parsedToken.Valid {
			abortWithError(c, http.StatusUnauthorized, fmt.Sprintf("%s: %v", errInvalidToken, err))
			return
		}

		// Guardar el token y claims en el contexto
		c.Set(cfg.ContextKey, parsedToken)
		c.Set(GetClaimsKey(cfg.ContextKey), parsedToken.Claims)

		c.Next()
	}
}

/* ---------------------------------------------------------------------------
   4) Helpers para extraer claims del contexto
   --------------------------------------------------------------------------- */

// ExtractClaim recupera un claim concreto (claimKey) del token almacenado
// en el contexto bajo la clave contextKey (por defecto, "token").
func ExtractClaim(c *gin.Context, claimKey, contextKey string) (string, error) {
	// Si no se especificó contextKey, usamos el default
	if contextKey == "" {
		contextKey = DefaultContextKey
	}

	// 1. Obtener el token del contexto
	tokenInterface, exists := c.Get(contextKey)
	if !exists {
		return "", fmt.Errorf(errTokenNotFound)
	}

	token, ok := tokenInterface.(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("invalid token type in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims type")
	}

	// 2. Buscar el claim específico
	claim, exists := claims[claimKey]
	if !exists {
		return "", fmt.Errorf("%s: %s", errClaimNotFound, claimKey)
	}

	return formatClaimValue(claim)
}

// ExtractUserID es un ejemplo que busca el claim "user_id" en el token
func ExtractUserID(c *gin.Context) (string, error) {
	return ExtractClaim(c, "user_id", "")
}

// GetClaimsKey retorna "contextKey_claims", usado para almacenar claims
// de forma paralela al token. Ej: "token_claims" si contextKey="token".
func GetClaimsKey(tokenKey string) string {
	if tokenKey == "" {
		tokenKey = DefaultContextKey
	}
	return tokenKey + ClaimsSuffix
}

/* ---------------------------------------------------------------------------
   5) Funciones auxiliares privadas: extracción de token, parse de claves, etc.
   --------------------------------------------------------------------------- */

// extrae el token del header o query, según cfg.TokenLookup (ej: "header:Authorization").
func extractToken(c *gin.Context, cfg Config) (string, error) {
	parts := strings.Split(cfg.TokenLookup, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf(errInvalidTokenLookup)
	}

	switch parts[0] {
	case "header":
		return extractFromHeader(c, parts[1], cfg.TokenPrefix)
	case "query":
		return extractFromQuery(c, parts[1])
	default:
		return "", fmt.Errorf(errUnsupportedLookupMethod)
	}
}

// extrae el token del header con un posible prefijo (ej. "Bearer ").
func extractFromHeader(c *gin.Context, header, prefix string) (string, error) {
	auth := c.GetHeader(header)
	if auth == "" {
		return "", fmt.Errorf(errMissingAuthHeader)
	}
	if prefix != "" && !strings.HasPrefix(auth, prefix) {
		return "", fmt.Errorf(errBearerPrefixRequired)
	}
	return strings.TrimPrefix(auth, prefix), nil
}

// extrae el token de un query param, ej. "?token=..."
func extractFromQuery(c *gin.Context, param string) (string, error) {
	token := c.Query(param)
	if token == "" {
		return "", fmt.Errorf(errMissingAuthHeader)
	}
	return token, nil
}

// decide si usar SecretKey (HMAC) o PublicKey (RSA) según el método de firma usado en el token.
func selectKeyFunc(token *jwt.Token, secretKey string, rsaKey *rsa.PublicKey) jwt.Keyfunc {
	switch token.Method.(type) {
	case *jwt.SigningMethodHMAC:
		if secretKey == "" {
			return nil
		}
		return func(token *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		}
	case *jwt.SigningMethodRSA:
		if rsaKey == nil {
			return nil
		}
		return func(token *jwt.Token) (any, error) {
			return rsaKey, nil
		}
	default:
		return nil
	}
}

// parsea la clave pública RSA desde un string PEM.
func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaKey, nil
}

// convierte el valor de un claim a string (si es posible).
func formatClaimValue(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		return fmt.Sprintf("%.0f", val), nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	case nil:
		return "", fmt.Errorf("%s: claim is nil", errInvalidClaimType)
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// Envía un JSON de error y corta la cadena de middlewares.
func abortWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
	c.Abort()
}
