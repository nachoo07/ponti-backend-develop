package pkgrabbit

import (
	"fmt"
	"os"
	"strconv"
)

// Bootstrap crea una instancia de Service (broker) usando la configuración obtenida de las variables de entorno.
func Bootstrap() (Service, error) {
	port, err := strconv.Atoi(os.Getenv("RABBITMQ_PORT"))
	if err != nil {
		return nil, fmt.Errorf("RABBITMQ_PORT no es un número válido: %w", err)
	}

	config := newConfig(
		os.Getenv("RABBITMQ_HOST"),
		port,
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_VHOST"),
		os.Getenv("RABBITMQ_QUEUE"),
		os.Getenv("RABBITMQ_EXCHANGE"),
		os.Getenv("RABBITMQ_EXCHANGE_TYPE"),
		os.Getenv("RABBITMQ_ROUTING_KEY"),
		os.Getenv("RABBITMQ_AUTO_ACK") == "true",
		os.Getenv("RABBITMQ_EXCLUSIVE") == "true",
		os.Getenv("RABBITMQ_NO_LOCAL") == "true",
		os.Getenv("RABBITMQ_NO_WAIT") == "true",
	)
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return newService(config)
}
