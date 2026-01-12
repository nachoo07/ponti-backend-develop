package pkgxaouth2

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	pkgoauth2 "github.com/alphacodinggroup/ponti-backend/pkg/authe/oauth2"
)

// BootstrapStd inicializa un service OAuth2 genérico
// usando golang.org/x/oauth2 con endpoints configurables.
func BootstrapStd(
	clientID, clientSecret, authURL, tokenURL, redirectURL string,
	scopes []string,
	timeoutSeconds int,
) (pkgoauth2.Service, error) {
	// 1) Leer variables de entorno si faltan
	if clientID == "" {
		clientID = os.Getenv("OAUTH2_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("OAUTH2_CLIENT_SECRET")
	}
	if authURL == "" {
		authURL = os.Getenv("OAUTH2_AUTH_URL")
	}
	if tokenURL == "" {
		tokenURL = os.Getenv("OAUTH2_TOKEN_URL")
	}
	if redirectURL == "" {
		redirectURL = os.Getenv("OAUTH2_REDIRECT_URL")
	}

	// 2) Scopes
	if len(scopes) == 0 {
		envScopes := os.Getenv("OAUTH2_SCOPES")
		if envScopes != "" {
			scopes = strings.Split(envScopes, ",")
		} else {
			scopes = []string{"profile"}
		}
	}

	// 3) Timeout
	if timeoutSeconds <= 0 {
		ts, _ := strconv.Atoi(os.Getenv("OAUTH2_TIMEOUT_SECONDS"))
		if ts > 0 {
			timeoutSeconds = ts
		} else {
			timeoutSeconds = 10
		}
	}

	// 4) Construir config embebiendo la BaseConfig
	cfg := &Config{
		BaseConfig: pkgoauth2.BaseConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			AuthURL:      authURL,
			TokenURL:     tokenURL,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			TimeoutSec:   timeoutSeconds,
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OAuth2 config: %w", err)
	}

	// 5) Crear service
	return NewService(cfg)
}
