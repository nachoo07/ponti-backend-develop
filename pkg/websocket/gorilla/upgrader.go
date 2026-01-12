package pkgws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader es la implementación por defecto usando Gorilla WebSocket.
type upgrader struct {
	upgrader *websocket.Upgrader
}

// NewUpgrader crea una nueva instancia de Upgrader usando la configuración proporcionada.
func newUpgrader(cfg Config) Upgrader {
	return &upgrader{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  cfg.GetReadBufferSize(),
			WriteBufferSize: cfg.GetWriteBufferSize(),
			// En producción, ajustar CheckOrigin según sea necesario.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Upgrade actualiza la conexión HTTP a una conexión WebSocket.
func (g *upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error) {
	c, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &conn{c}, nil
}
