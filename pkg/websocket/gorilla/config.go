package pkgws

import "errors"

// Config define la configuración para el servidor WebSocket.
type config struct {
	readBufferSize  int
	writeBufferSize int
}

// NewConfig crea una nueva configuración utilizando los parámetros proporcionados y la valida.
// Retorna un error si la validación falla.
func newConfig(readBufferSize, writeBufferSize int) Config {
	cfg := &config{
		readBufferSize:  readBufferSize,
		writeBufferSize: writeBufferSize,
	}
	return cfg
}

// GetReadBufferSize retorna el tamaño del buffer de lectura.
func (cfg *config) GetReadBufferSize() int {
	return cfg.readBufferSize
}

// GetWriteBufferSize retorna el tamaño del buffer de escritura.
func (cfg *config) GetWriteBufferSize() int {
	return cfg.writeBufferSize
}

// Validate verifica que los valores de configuración sean válidos.
func (cfg *config) Validate() error {
	if cfg.readBufferSize <= 0 {
		return errors.New("readBufferSize must be greater than 0")
	}
	if cfg.writeBufferSize <= 0 {
		return errors.New("writeBufferSize must be greater than 0")
	}
	return nil
}
