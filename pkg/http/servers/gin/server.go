package pkggin

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// con sigleton
var (
	instance  Server
	once      sync.Once
	initError error
)

type server struct {
	router *gin.Engine
	config Config
}

func newServer(config Config) (Server, error) {
	once.Do(func() {
		err := config.Validate()
		if err != nil {
			initError = err
			return
		}

		r := gin.New()
		instance = &server{
			config: config,
			router: r,
		}
	})
	return instance, initError
}

// sin singleton
// type server struct {
// 	router *gin.Engine
// 	config Config
// }

// func newServer(cfg Config) (Server, error) {
// 	r := gin.Default()
// 	return &server{
// 		router: r,
// 		config: cfg,
// 	}, nil
// }

func newTestServer() (Server, error) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	testConfig := &config{
		routerPort: "8080",
		apiVersion: "v1",
	}

	return &server{
		router: r,
		config: testConfig,
	}, nil
}

// RunServer lanza el servidor en el puerto configurado.
func (s *server) RunServer(ctx context.Context) error {
	// Ejemplo de "Run" bloqueante:
	return s.router.Run(":" + s.config.GetRouterPort())
}

// GetRouter expone el router para poder añadir rutas, middlewares, etc.
func (s *server) GetRouter() *gin.Engine {
	return s.router
}

// GetApiVersion retorna la versión configurada.
func (s *server) GetApiVersion() string {
	return s.config.GetApiVersion()
}

// WrapH sirve para anidar un http.Handler dentro de Gin.
func (s *server) WrapH(h http.Handler) gin.HandlerFunc {
	return gin.WrapH(h)
}
