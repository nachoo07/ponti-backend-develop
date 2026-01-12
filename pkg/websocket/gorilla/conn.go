package pkgws

import "github.com/gorilla/websocket"

// conn es un wrapper de *websocket.Conn que implementa Conn.
type conn struct {
	*websocket.Conn
}

// ReadMessage lee un mensaje de la conexión.
func (c *conn) ReadMessage() (int, []byte, error) {
	return c.Conn.ReadMessage()
}

// WriteMessage escribe un mensaje en la conexión.
func (c *conn) WriteMessage(messageType int, data []byte) error {
	return c.Conn.WriteMessage(messageType, data)
}

// Close cierra la conexión.
func (c *conn) Close() error {
	return c.Conn.Close()
}
