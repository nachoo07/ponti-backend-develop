package pkgsession

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
)

var (
	instance SessionManager
	once     sync.Once
)

type sessionManager struct {
	store *sessions.CookieStore
}

// newSessionManager es una función que crea una única instancia del manejador de sesiones
func newSessionManager(c Config) (SessionManager, error) {
	var err error
	once.Do(func() {
		// Crear el almacén de cookies
		store := sessions.NewCookieStore([]byte(c.GetSecretKey()))

		// Comprobación opcional de algún error durante la inicialización
		if store == nil {
			err = fmt.Errorf("failed to create session store")
			return
		}

		instance = &sessionManager{
			store: store,
		}
	})

	// Retornar la instancia y el error si ocurrió
	return instance, err
}

// Implementación de los métodos definidos en la interfaz SessionManager
func (r *sessionManager) Get(rq *http.Request, name string) (*sessions.Session, error) {
	return r.store.Get(rq, name)
}

func (r *sessionManager) Save(rq *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return session.Save(rq, w)
}

func (r *sessionManager) New(rq *http.Request, name string) (*sessions.Session, error) {
	return r.store.New(rq, name)
}
