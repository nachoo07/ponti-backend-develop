package pkgrabbit

import (
	"fmt"
	"os"
	"strconv"
)

// BootstrapConsumer crea una instancia de Consumer usando la configuración obtenida de variables de entorno.
func BootstrapConsumer() (Consumer, error) {
	// Obtener variables de entorno.
	host := os.Getenv("RABBITMQ_HOST")
	portStr := os.Getenv("RABBITMQ_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("RABBITMQ_PORT is not a valid number: %w", err)
	}
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	vhost := os.Getenv("RABBITMQ_VHOST")
	exchange := os.Getenv("RABBITMQ_EXCHANGE")
	exchangeType := os.Getenv("RABBITMQ_EXCHANGE_TYPE")
	durable, err := strconv.ParseBool(os.Getenv("RABBITMQ_DURABLE"))
	if err != nil {
		durable = false
	}
	autoDelete, err := strconv.ParseBool(os.Getenv("RABBITMQ_AUTO_DELETE"))
	if err != nil {
		autoDelete = false
	}
	internal, err := strconv.ParseBool(os.Getenv("RABBITMQ_INTERNAL"))
	if err != nil {
		internal = false
	}
	noWait, err := strconv.ParseBool(os.Getenv("RABBITMQ_NO_WAIT"))
	if err != nil {
		noWait = false
	}
	confirmBufferSize := 10
	if bufStr := os.Getenv("RABBITMQ_CONFIRM_BUFFER_SIZE"); bufStr != "" {
		if buf, err := strconv.Atoi(bufStr); err == nil {
			confirmBufferSize = buf
		}
	}

	// Crear la configuración.
	config := newConfig(
		host,
		port,
		user,
		password,
		vhost,
		exchange,
		exchangeType,
		durable,
		autoDelete,
		internal,
		noWait,
		confirmBufferSize,
	)
	return NewConsumer(config, nil)
}
