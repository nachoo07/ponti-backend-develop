package pkcresty

import (
	"time"
)

func Bootstrap(baseURL string, timeout time.Duration) (Client, error) {
	// Crear la configuración del cliente
	cfg := newConfig(baseURL, timeout)

	// Validar configuración
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Retornar nuevo cliente
	return newClient(cfg)
}
