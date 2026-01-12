package pkgjwt

import (
	"fmt"
	"log"
	"time"
)

type config struct {
	secret                   string
	accessExpirationMinutes  int
	refreshExpirationMinutes int
}

// newConfig crea una nueva configuración de JWT
func newConfig(secretKey string, accessExpirationMinutes, refreshExpirationMinutes int) Config {
	return &config{
		secret:                   secretKey,
		accessExpirationMinutes:  accessExpirationMinutes,
		refreshExpirationMinutes: refreshExpirationMinutes,
	}
}

// GetSecretKey devuelve la clave secreta para firmar los tokens JWT
func (c *config) GetSecretKey() string {
	return c.secret
}

// GetAccessExpiration devuelve la duración de expiración del token de acceso
func (c *config) GetAccessExpiration() time.Duration {
	return time.Duration(c.accessExpirationMinutes) * time.Minute
}

// GetRefreshExpiration devuelve la duración de expiración del token de refresco
func (c *config) GetRefreshExpiration() time.Duration {
	return time.Duration(c.refreshExpirationMinutes) * time.Minute
}

func (c *config) Validate() error {
	if c.secret == "" {
		return fmt.Errorf("JWT secret not configured")
	}
	if c.accessExpirationMinutes <= 0 {
		log.Printf("WARNING: Default JWT access expiration not configured or must be greater than 0")
	}
	if c.refreshExpirationMinutes <= 0 {
		log.Printf("WARNING: Default JWT refresh expiration not configuredo or must be greater than 0")
	}

	return nil
}
