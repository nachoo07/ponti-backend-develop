package pkgredis

import (
	"fmt"
	"time"
)

// config estructura de configuración de Redis
type config struct {
	Address           string
	Password          string
	DB                int
	DefaultExpiration time.Duration // Parámetro de expiración predeterminada
}

// newConfig crea una nueva configuración de Redis
func newConfig(address, password string, db int) *config {
	return &config{
		Address:  address,
		Password: password,
		DB:       db,
		// DefaultExpiration: time.Minute, // Ejemplo si quieres setear un valor por defecto
	}
}

// Validate verifica que la configuración de Redis sea válida
func (c *config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("REDIS_ADDRESS is required")
	}
	if c.DB < 0 {
		return fmt.Errorf("REDIS_DB must be a non-negative integer")
	}
	return nil
}

// GetAddress devuelve la dirección de Redis
func (c *config) GetAddress() string {
	return c.Address
}

// GetPassword devuelve la contraseña de Redis
func (c *config) GetPassword() string {
	return c.Password
}

// GetDB devuelve el número de la base de datos de Redis
func (c *config) GetDB() int {
	return c.DB
}

// GetDefaultExpiration devuelve la expiración por defecto
func (c *config) GetDefaultExpiration() time.Duration {
	return c.DefaultExpiration
}
