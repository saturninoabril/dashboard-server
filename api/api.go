package api

import "github.com/gorilla/mux"

const (
	pathPrefix = "/api/v1"
)

// Register registers the API endpoints on the given router.
func Register(public *mux.Router, context *Context) {
	initializePublicRoutes(public, context)
}

// initializePublicRoutes instantiates routes that listen on the public interface
func initializePublicRoutes(rootRouter *mux.Router, context *Context) {
	apiRouter := rootRouter.PathPrefix(pathPrefix).Subrouter()

	initHealth(apiRouter, context)
	initUser(apiRouter, context)
	initOAuth(apiRouter, context)
}
