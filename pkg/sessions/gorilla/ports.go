package pkgsession

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionManager define la interfaz para el manejo de sesiones
type SessionManager interface {
	Get(*http.Request, string) (*sessions.Session, error)
	Save(*http.Request, http.ResponseWriter, *sessions.Session) error
	New(*http.Request, string) (*sessions.Session, error)
}

// Config define la interfaz de la configuración de sesiones
type Config interface {
	GetSecretKey() string
	Validate() error
}
