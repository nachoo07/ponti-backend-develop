package pkgws

import (
	"os"
	"strconv"
)

// Bootstrap inicializa y retorna un Upgrader configurado.
// Lee las variables de entorno y las pasa como argumentos a NewConfig.
func Bootstrap() (Upgrader, error) {
	// Valores por defecto
	readBufferSize := 1024
	writeBufferSize := 1024

	// Leer WS_READ_BUFFER_SIZE
	if rbs := os.Getenv("WS_READ_BUFFER_SIZE"); rbs != "" {
		if v, err := strconv.Atoi(rbs); err == nil {
			readBufferSize = v
		}
	}

	// Leer WS_WRITE_BUFFER_SIZE
	if wbs := os.Getenv("WS_WRITE_BUFFER_SIZE"); wbs != "" {
		if v, err := strconv.Atoi(wbs); err == nil {
			writeBufferSize = v
		}
	}

	config := newConfig(
		readBufferSize,
		writeBufferSize,
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newUpgrader(config), nil
}
