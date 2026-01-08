package pkggorhttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type server struct {
	router     *mux.Router
	config     Config
	httpServer *http.Server
}

var (
	instance  Server
	once      sync.Once
	initError error
)

// NewServer inicializa (usando el patrón singleton) el servidor con la configuración proporcionada.
func newServer(config Config) (Server, error) {
	once.Do(func() {
		if err := config.Validate(); err != nil {
			initError = err
			return
		}

		// Crear el router de Gorilla Mux.
		router := mux.NewRouter()

		// Registro de rutas (por ejemplo, un endpoint de salud).
		router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}).Methods("GET")

		// Aplicar middlewares: Logging (y se puede agregar Recovery u otros según se requiera).
		loggedRouter := handlers.LoggingHandler(os.Stdout, router)

		// Configurar el servidor HTTP con timeouts adecuados.
		httpSrv := &http.Server{
			Addr:         ":" + config.GetPort(),
			Handler:      loggedRouter,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		instance = &server{
			router:     router,
			config:     config,
			httpServer: httpSrv,
		}
	})
	return instance, initError
}

// Run inicia el servidor HTTP y espera a que se cancele el contexto para realizar un shutdown controlado.
func (s *server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("error durante el shutdown: %w", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *server) GetAPIVersion() string {
	return s.config.GetAPIVersion()
}

func (s *server) GetHandler() http.Handler {
	return s.httpServer.Handler
}

func (s *server) GetRouter() *mux.Router {
	return s.router
}
