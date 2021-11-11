package app

import (
	"context"
	"net/url"
	"path"

	"github.com/google/go-github/github"
	"github.com/saturninoabril/dashboard-server/model"
	"golang.org/x/oauth2"
)

func (a *App) CreateOAuthState() (*model.OAuthState, error) {
	return a.store.OAuthState().CreateOAuthState()
}

func (a *App) GetOAuthState(idOrToken string) (*model.OAuthState, error) {
	return a.store.OAuthState().GetOAuthState(idOrToken)
}

func (a *App) GetOAuthConfig() *oauth2.Config {
	scopes := []string{
		string(github.ScopeUserEmail),
		string(github.ScopeReadOrg),
	}

	baseURL := "https://github.com/"
	authURL, _ := url.Parse(baseURL)
	tokenURL, _ := url.Parse(baseURL)

	authURL.Path = path.Join(authURL.Path, "login", "oauth", "authorize")
	tokenURL.Path = path.Join(tokenURL.Path, "login", "oauth", "access_token")

	return &oauth2.Config{
		ClientID:     a.config.GithubOAuth.ClientID,
		ClientSecret: a.config.GithubOAuth.ClientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL.String(),
			TokenURL:  tokenURL.String(),
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
}

func (a *App) GithubConnectToken(token oauth2.Token) (*github.Client, error) {
	client, err := GetGitHubClient(token)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetGitHubClient(token oauth2.Token) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(&token)
	authenticatedClient := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(authenticatedClient), nil
}
