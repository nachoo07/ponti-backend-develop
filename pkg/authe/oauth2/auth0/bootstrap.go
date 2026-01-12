package pkgauth0

import (
	"fmt"
	"os"
	"strconv"

	pkgoauth2 "github.com/alphacodinggroup/ponti-backend/pkg/authe/oauth2"
	// Importas tu paquete base con interfaces/structs
	// Este paquete define tu Config y NewService (service.go)
)

// BootstrapAuth0 inicializa la configuración y crea el service de Auth0.
// Lee parámetros de función o variables de entorno (AUTH0_DOMAIN, AUTH0_AUDIENCE, etc.).
func Bootstrap(domain, clientID, clientSecret, audience string, timeoutSeconds int) (pkgoauth2.Service, error) {

	// 1) Leer variables de entorno si faltan
	if domain == "" {
		domain = os.Getenv("AUTH0_DOMAIN")
	}
	if clientID == "" {
		clientID = os.Getenv("AUTH0_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	}
	if audience == "" {
		audience = os.Getenv("AUTH0_AUDIENCE")
	}

	// 2) Timeout
	if timeoutSeconds <= 0 {
		ts, _ := strconv.Atoi(os.Getenv("AUTH0_TIMEOUT_SECONDS"))
		if ts > 0 {
			timeoutSeconds = ts
		} else {
			timeoutSeconds = 10
		}
	}

	// 3) Crear config específica de Auth0 (ej. struct Config embed de pkgoauth2.BaseConfig)
	cfg := &Config{
		BaseConfig: pkgoauth2.BaseConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			// Aunque en Auth0 se calculan, los dejamos para cumplir la interfaz base
			AuthURL:     "https://" + domain + "/authorize",
			TokenURL:    "https://" + domain + "/oauth/token",
			RedirectURL: "",                            // si no se usa Authorization Code, podría quedar vacío
			Scopes:      []string{"openid", "profile"}, // o lo que necesites
			TimeoutSec:  timeoutSeconds,
		},
		Domain:   domain,
		Audience: audience,
	}

	// 4) Validar la config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Auth0 config: %w", err)
	}

	// 5) Crear el service Auth0
	return NewService(cfg)
}
