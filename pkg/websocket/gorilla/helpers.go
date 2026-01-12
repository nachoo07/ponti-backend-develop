package pkgws

import (
	"net/http"
)

// HandleWebSocket es una función helper que realiza el upgrade de la conexión
// y delega el manejo al handler proporcionado.
// NOTA: Esta función no crea el servidor HTTP, solo procesa la conexión WebSocket.
func HandleWebSocket(w http.ResponseWriter, r *http.Request, upgrader Upgrader, handler func(Conn)) {
	conn, err := upgrader.Upgrade(w, r)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}
	go func() {
		defer conn.Close()
		handler(conn)
	}()
}
