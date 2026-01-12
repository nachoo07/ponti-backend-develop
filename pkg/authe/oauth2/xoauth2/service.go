package pkgxaouth2

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	pkgoauth2 "github.com/alphacodinggroup/ponti-backend/pkg/authe/oauth2"
)

type service struct {
	cfg       *Config
	oauth2Cfg *oauth2.Config
	timeout   time.Duration
}

func NewService(cfg *Config) (pkgoauth2.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OAuth2 configuration: %w", err)
	}

	oCfg := &oauth2.Config{
		ClientID:     cfg.GetClientID(),
		ClientSecret: cfg.GetClientSecret(),
		Scopes:       cfg.GetScopes(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.GetAuthURL(),
			TokenURL: cfg.GetTokenURL(),
		},
		RedirectURL: cfg.GetRedirectURL(),
	}

	return &service{
		cfg:       cfg,
		oauth2Cfg: oCfg,
		timeout:   cfg.GetTimeout(),
	}, nil
}

func (s *service) GetAuthCodeURL(state string) string {
	return s.oauth2Cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *service) ExchangeCode(ctx context.Context, code string) (*pkgoauth2.OAuth2Token, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	token, err := s.oauth2Cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return &pkgoauth2.OAuth2Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*pkgoauth2.OAuth2Token, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ts := s.oauth2Cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	newToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &pkgoauth2.OAuth2Token{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		Expiry:       newToken.Expiry,
	}, nil
}

func (s *service) ValidateToken(ctx context.Context, tokenStr string) (*pkgoauth2.TokenClaims, error) {
	// Aquí podrías hacer introspección, decodificar JWT, etc.
	return nil, fmt.Errorf("ValidateToken not implemented for pkgxaouth2")
}
