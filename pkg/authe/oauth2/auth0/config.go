package pkgauth0

import (
	"fmt"
	"time"

	pkgoauth2 "github.com/alphacodinggroup/ponti-backend/pkg/authe/oauth2"
)

// Config implementa pkgoauth2.Config para Auth0
type Config struct {
	pkgoauth2.BaseConfig

	// Campos específicos de Auth0
	Domain   string
	Audience string
}

// Validate extiende la validación base
func (c *Config) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	if c.Domain == "" {
		return fmt.Errorf("auth0 domain is required")
	}
	if c.Audience == "" {
		return fmt.Errorf("auth0 audience is required")
	}
	return nil
}

// Sobrescribimos o ajustamos GetAuthURL, GetTokenURL si necesitamos
func (c *Config) GetAuthURL() string {
	// Por ejemplo, si Auth0 construye el AuthURL a partir del Domain
	return "https://" + c.Domain + "/authorize"
}

func (c *Config) GetTokenURL() string {
	// Auth0 token endpoint
	return "https://" + c.Domain + "/oauth/token"
}

func (c *Config) GetTimeout() time.Duration {
	if c.TimeoutSec <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.TimeoutSec) * time.Second
}
