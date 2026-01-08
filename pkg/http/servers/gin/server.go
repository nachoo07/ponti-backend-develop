package pkggin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	instance  *Server
	once      sync.Once
	initError error
)

type ConfigPort interface {
	GetRouterPort() string
	Validate() error
}

type Server struct {
	router *gin.Engine
	config ConfigPort
}

func newServer(config ConfigPort) (*Server, error) {
	once.Do(func() {
		err := config.Validate()
		if err != nil {
			initError = err
			return
		}

		r := gin.New()

		r.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
			})
		})

		r.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		instance = &Server{
			config: config,
			router: r,
		}
	})
	return instance, initError
}

func newTestServer() (*Server, error) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	testConfig := &Config{
		routerPort: "8080",
		apiVersion: "v1",
	}

	return &Server{
		router: r,
		config: testConfig,
	}, nil
}

func (s *Server) RunServer(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:    ":" + s.config.GetRouterPort(),
		Handler: s.router,
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("Context canceled. Shutting down HTTP Server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP Server shutdown failed: %w", err)
		}
		return nil

	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP Server error: %w", err)
		}
		return nil
	}
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) WrapH(h http.Handler) gin.HandlerFunc {
	return gin.WrapH(h)
}
