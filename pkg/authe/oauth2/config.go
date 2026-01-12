package pkgoauth2

import "time"

// BaseConfig es una implementación mínima de Config.
// Puedes usarla tal cual o embederla en otra struct.
type BaseConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string
	Scopes       []string
	TimeoutSec   int
}

// Validate revisa campos básicos. Puedes agregar checks según tu necesidad.
func (c *BaseConfig) Validate() error {
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if c.ClientSecret == "" {
		return ErrMissingClientSecret
	}
	if c.AuthURL == "" {
		return ErrMissingAuthURL
	}
	if c.TokenURL == "" {
		return ErrMissingTokenURL
	}
	if c.RedirectURL == "" {
		return ErrMissingRedirectURL
	}
	if len(c.Scopes) == 0 {
		return ErrMissingScopes
	}
	return nil
}

func (c *BaseConfig) GetClientID() string     { return c.ClientID }
func (c *BaseConfig) GetClientSecret() string { return c.ClientSecret }
func (c *BaseConfig) GetAuthURL() string      { return c.AuthURL }
func (c *BaseConfig) GetTokenURL() string     { return c.TokenURL }
func (c *BaseConfig) GetRedirectURL() string  { return c.RedirectURL }
func (c *BaseConfig) GetScopes() []string     { return c.Scopes }

func (c *BaseConfig) GetTimeout() time.Duration {
	if c.TimeoutSec <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.TimeoutSec) * time.Second
}

// Posibles errores específicos
var (
	ErrMissingClientID     = Error("missing client_id")
	ErrMissingClientSecret = Error("missing client_secret")
	ErrMissingAuthURL      = Error("missing auth_url")
	ErrMissingTokenURL     = Error("missing token_url")
	ErrMissingRedirectURL  = Error("missing redirect_url")
	ErrMissingScopes       = Error("missing scopes")
)

// Error tipo para simplificar
type Error string

func (e Error) Error() string { return string(e) }
