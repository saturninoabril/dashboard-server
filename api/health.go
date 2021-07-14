package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// initHealth registers status endpoints on the given router.
func initHealth(apiRouter *mux.Router, context *Context) {
	healthRouter := apiRouter.PathPrefix("/health").Subrouter()
	healthRouter.Handle("", newAPIHandler(context, handleHealthCheck)).Methods("GET")
}

// handleHealthCheck responds to GET /api/v1/health as a health check.
func handleHealthCheck(c *Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
