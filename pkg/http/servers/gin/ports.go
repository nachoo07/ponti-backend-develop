package pkggin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server expone las operaciones principales de tu servidor.
type Server interface {
	RunServer(context.Context) error
	GetApiVersion() string
	GetRouter() *gin.Engine
	WrapH(h http.Handler) gin.HandlerFunc
}

type Config interface {
	GetRouterPort() string
	SetRouterPort(string)
	GetApiVersion() string
	SetApiVersion(string)
	Validate() error
}
