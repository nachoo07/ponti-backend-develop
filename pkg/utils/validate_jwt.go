package pkgutils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// DefaultContextKey es la clave por defecto para almacenar el token en el contexto.
	DefaultContextKey = "token"
	// ClaimsSuffix se añade a la clave del token para formar la clave de los claims.
	ClaimsSuffix = "_claims"
	// authHeaderName define el nombre del header donde se espera la autorización.
	authHeaderName = "Authorization"
	// bearerPrefix es el prefijo esperado en el header de autorización.
	bearerPrefix = "Bearer "
	// errMissingAuth es el mensaje de error cuando falta el header de autorización.
	errMissingAuth = "authorization header required"
	// errInvalidLookup se utiliza cuando la configuración para extraer el token es inválida.
	errInvalidLookup = "invalid token lookup config"
	// errUnsupported se utiliza cuando el método para extraer el token no es soportado.
	errUnsupported = "unsupported token lookup method"
)

// Config define la configuración común para la validación y extracción de JWT.
type Config struct {
	SecretKey    string // Clave secreta para tokens firmados con HMAC.
	PublicKeyPEM string // Cadena en formato PEM para la clave pública RSA.
	TokenLookup  string // Define cómo y desde dónde extraer el token (ej. "header:Authorization" o "query:token").
	TokenPrefix  string // Prefijo a remover del token (ej. "Bearer ").
	ContextKey   string // Clave para almacenar el token en el contexto de la request.
}

// NewConfigFromEnv crea una instancia de Config leyendo las variables de entorno,
// utilizando valores por defecto cuando no se proporcionen.
func NewConfigFromEnv() Config {
	return Config{
		SecretKey:    os.Getenv("JWT_SECRET_KEY"),
		PublicKeyPEM: os.Getenv("JWT_PUBLIC_PEM_KEY"),
		// Si no se define la variable de entorno, se utiliza el default "header:Authorization".
		TokenLookup: getEnvOrDefault("JWT_TOKEN_LOOKUP", "header:"+authHeaderName),
		// Si no se define la variable, se utiliza "Bearer " como prefijo.
		TokenPrefix: getEnvOrDefault("JWT_TOKEN_PREFIX", bearerPrefix),
		// Si no se define la variable, se utiliza "token" como clave en el contexto.
		ContextKey: getEnvOrDefault("JWT_CONTEXT_KEY", DefaultContextKey),
	}
}

// ParseRSAPublicKey convierte una cadena PEM en una clave pública RSA.
// Esta función es útil cuando se usan tokens firmados con el algoritmo RSA.
func ParseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	// Decodifica la cadena PEM a un bloque de datos.
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	// Parsear la clave pública a partir de los bytes del bloque.
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// Asegurarse de que la clave parseada es de tipo RSA.
	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaKey, nil
}

// ExtractTokenFromRequest extrae el token JWT de la request HTTP según la configuración especificada.
// Soporta extracción desde header o query string.
func ExtractTokenFromRequest(r *http.Request, cfg Config) (string, error) {
	// Se espera que TokenLookup tenga el formato "origen:clave", por ejemplo "header:Authorization"
	parts := strings.Split(cfg.TokenLookup, ":")
	if len(parts) != 2 {
		return "", errors.New(errInvalidLookup)
	}
	switch parts[0] {
	case "header":
		// Extrae el token del header especificado y remueve el prefijo.
		return extractFromHeader(r, parts[1], cfg.TokenPrefix)
	case "query":
		// Extrae el token desde los parámetros de la URL.
		return extractFromQuery(r, parts[1])
	default:
		// Si el método no es soportado se retorna un error.
		return "", errors.New(errUnsupported)
	}
}

// SelectKeyFunc determina la función para obtener la clave de verificación según el método de firma del token.
// Dependiendo de si es HMAC o RSA se retornará la clave correspondiente.
func SelectKeyFunc(token *jwt.Token, secretKey string, rsaKey *rsa.PublicKey) jwt.Keyfunc {
	switch token.Method.(type) {
	// Si se usa HMAC, se requiere la clave secreta.
	case *jwt.SigningMethodHMAC:
		if secretKey == "" {
			return nil
		}
		return func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		}
	// Si se usa RSA, se requiere la clave pública.
	case *jwt.SigningMethodRSA:
		if rsaKey == nil {
			return nil
		}
		return func(token *jwt.Token) (interface{}, error) {
			return rsaKey, nil
		}
	default:
		// Para otros métodos no soportados se retorna nil.
		return nil
	}
}

// GetClaimsKey genera la clave para almacenar los claims del token en el contexto.
// Se concatena la clave base con un sufijo.
func GetClaimsKey(tokenKey string) string {
	if tokenKey == "" {
		tokenKey = DefaultContextKey
	}
	return tokenKey + ClaimsSuffix
}

// ExtractClaim extrae el valor de un claim específico de un token JWT.
// Retorna un error si el claim no existe o el tipo de claims no es el esperado.
func ExtractClaim(token *jwt.Token, claimKey string) (string, error) {
	// Se espera que los claims sean del tipo jwt.MapClaims.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims type")
	}
	value, exists := claims[claimKey]
	if !exists {
		return "", fmt.Errorf("claim %s not found", claimKey)
	}
	// Se formatea el valor del claim a string para facilitar su uso.
	return formatClaimValue(value)
}

//
// Funciones Helper
//

// getEnvOrDefault retorna el valor de una variable de entorno o, si no está definida, el valor por defecto especificado.
func getEnvOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// extractFromHeader obtiene el token desde el header HTTP especificado.
// Verifica que el header exista y que comience con el prefijo esperado.
func extractFromHeader(r *http.Request, header, prefix string) (string, error) {
	auth := r.Header.Get(header)
	if auth == "" {
		return "", errors.New(errMissingAuth)
	}
	// Comprueba que el valor del header inicia con el prefijo (por ejemplo, "Bearer ").
	if prefix != "" && !strings.HasPrefix(auth, prefix) {
		return "", fmt.Errorf("authorization header must start with %s", prefix)
	}
	// Remueve el prefijo para retornar solo el token.
	return strings.TrimPrefix(auth, prefix), nil
}

// extractFromQuery obtiene el token desde los parámetros de la URL usando el nombre especificado.
func extractFromQuery(r *http.Request, param string) (string, error) {
	token := r.URL.Query().Get(param)
	if token == "" {
		return "", errors.New(errMissingAuth)
	}
	return token, nil
}

// formatClaimValue convierte un valor de claim a string.
// Soporta múltiples tipos comunes como string, float64, int, int64 y bool.
func formatClaimValue(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		// Se formatea sin decimales.
		return fmt.Sprintf("%.0f", val), nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	default:
		// Para otros tipos, se usa la representación por defecto.
		return fmt.Sprintf("%v", val), nil
	}
}
