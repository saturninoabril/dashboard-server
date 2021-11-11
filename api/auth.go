package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// initAuth registers OAuth endpoints on the given router.
func initOAuth(apiRouter *mux.Router, context *Context) {
	githubOAuthRouter := apiRouter.PathPrefix("/oauth/github").Subrouter()
	githubOAuthRouter.Handle("/connect", newAPIHandler(context, handleGithubConnect)).Methods(http.MethodGet)
	githubOAuthRouter.Handle("/complete", newAPIHandler(context, handleGithubComplete)).Methods(http.MethodPost)
}

// handleGithubConnect connects app to Github
// Responds to GET /api/v1/oauth/github/connect
func handleGithubConnect(c *Context, w http.ResponseWriter, r *http.Request) {
	state, err := c.App.CreateOAuthState()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("error creating oauth state"))
		return
	}

	conf := c.App.GetOAuthConfig()
	url := conf.AuthCodeURL(state.Token, oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, http.StatusFound)
}

// handleGithubComplete completes app to Github
// Responds to POST /api/v1/oauth/github/complete
func handleGithubComplete(c *Context, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("missing authorization code"))
		return
	}

	stateToken := r.URL.Query().Get("state")
	storedOAuthState, err := c.App.GetOAuthState(stateToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("invalid state"))
		return
	}

	if storedOAuthState != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.writeAndLogError(w, errors.New("state expired or not found"))
		return
	}

	ctx := context.Background()
	conf := c.App.GetOAuthConfig()

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.New("failed to exchange oauth code into token"))
		return
	}

	githubClient, err := c.App.GithubConnectToken(*tok)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.New("failed to authenticate with Github"))
		return
	}

	gitUser, _, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.writeAndLogError(w, errors.New("failed to get authenticated GitHub user"))
		return
	}

	b, _ := json.Marshal(gitUser)
	w.Write(b)
}
