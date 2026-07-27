package github

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type OAuthService struct {
	config *oauth2.Config
}

func NewOAuthService(cfg OAuthConfig) *OAuthService {
	return &OAuthService{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"read:user", "repo"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (s *OAuthService) AuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *OAuthService) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

func (s *OAuthService) OAuthConfig() *oauth2.Config {
	return s.config
}
