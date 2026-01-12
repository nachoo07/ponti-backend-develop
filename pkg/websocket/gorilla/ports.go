package pkgws

import "net/http"

// Conn abstrae una conexión WebSocket.
type Conn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Upgrader define la interfaz para actualizar una conexión HTTP a WebSocket.
type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error)
}

type Config interface {
	Validate() error
	GetReadBufferSize() int
	GetWriteBufferSize() int
}
