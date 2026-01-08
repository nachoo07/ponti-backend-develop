package pkcresty

import (
	"log"
	"time"
)

// config implementa la interfaz Config.
type config struct {
	baseURL string
	timeout time.Duration // timeout en segundos
}

// newConfig crea una nueva configuración para el cliente Resty.
func newConfig(baseURL string, timeout time.Duration) Config {
	return &config{
		baseURL: baseURL,
		timeout: timeout,
	}
}

// GetBaseURL retorna la URL base configurada.
func (c *config) GetBaseURL() string {
	return c.baseURL
}

// SetBaseURL establece la URL base.
func (c *config) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// GetTimeout retorna el timeout configurado en segundos.
func (c *config) GetTimeout() time.Duration {
	return c.timeout
}

// SetTimeout establece el timeout en segundos.
func (c *config) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Validate verifica que la configuración sea válida.
func (c *config) Validate() error {
	if c.baseURL == "" {
		log.Printf("WARNING: Default base URL is not configured")
	}
	if c.timeout <= 0 {
		log.Printf("WARNING: Default timeout or must be greater than 0")
	}
	return nil
}
