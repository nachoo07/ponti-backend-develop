package pkggorhttp

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Server interface {
	Run(ctx context.Context) error
	GetAPIVersion() string
	GetHandler() http.Handler
	GetRouter() *mux.Router
}

type Config interface {
	GetPort() string
	GetAPIVersion() string
	Validate() error
}
