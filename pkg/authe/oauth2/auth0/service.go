package pkgauth0

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	pkgoauth2 "github.com/alphacodinggroup/ponti-backend/pkg/authe/oauth2"
	"github.com/auth0-community/go-auth0"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/square/go-jose.v2"
)

type service struct {
	cfg          *Config
	jwkValidator *auth0.JWTValidator
	httpClient   *http.Client
}

// NewService crea un service para Auth0
func NewService(cfg *Config) (pkgoauth2.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Auth0 configuration: %w", err)
	}

	// Configurar validador JWK (Auth0)
	domainURL := "https://" + cfg.Domain + "/"
	jwksURI := domainURL + ".well-known/jwks.json"

	secretProvider := auth0.NewJWKClient(
		auth0.JWKClientOptions{URI: jwksURI},
		nil, // No pasamos extractor
	)

	configuration := auth0.NewConfiguration(
		secretProvider,
		[]string{cfg.Audience},
		domainURL,
		jose.RS256,
	)
	validator := auth0.NewValidator(configuration, nil)

	return &service{
		cfg:          cfg,
		jwkValidator: validator,
		httpClient:   &http.Client{Timeout: cfg.GetTimeout()},
	}, nil
}

// GetAuthCodeURL no es tan común en Auth0 + Client Credentials
// pero podríamos implementarlo si usamos Authorization Code Flow
func (s *service) GetAuthCodeURL(state string) string {
	// AuthURL se obtiene de s.cfg.GetAuthURL() si implementamos el flow
	return "not_implemented"
}

// ExchangeCode intercambia un código (no implementado en este ejemplo)
func (s *service) ExchangeCode(ctx context.Context, code string) (*pkgoauth2.OAuth2Token, error) {
	return nil, fmt.Errorf("ExchangeCode not implemented for Auth0 example")
}

// RefreshToken no implementado en este ejemplo
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*pkgoauth2.OAuth2Token, error) {
	return nil, fmt.Errorf("RefreshToken not implemented for Auth0 example")
}

// ValidateToken valida un token usando la librería go-auth0
func (s *service) ValidateToken(ctx context.Context, tokenStr string) (*pkgoauth2.TokenClaims, error) {
	// Creamos una request ficticia para que el validador extraiga el token del header
	req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	token, err := s.jwkValidator.ValidateRequest(req)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extraer claims
	var claims map[string]any
	if err := token.Claims(nil, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Convertir a TokenClaims genérico
	t := &pkgoauth2.TokenClaims{}
	if sub, ok := claims["sub"].(string); ok {
		t.Subject = sub
	}
	return t, nil
}

// GetClientCredentialsToken (ejemplo de Client Credentials Flow en Auth0)
func (s *service) GetClientCredentialsToken(ctx context.Context) (string, error) {
	domainURL := "https://" + s.cfg.Domain + "/"
	config := &clientcredentials.Config{
		ClientID:     s.cfg.GetClientID(),
		ClientSecret: s.cfg.GetClientSecret(),
		TokenURL:     domainURL + "oauth/token",
		// Scopes:       []string{"..."},
		EndpointParams: url.Values{
			"audience": {s.cfg.Audience},
		},
	}

	token, err := config.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get client credentials token: %w", err)
	}
	return token.AccessToken, nil
}
