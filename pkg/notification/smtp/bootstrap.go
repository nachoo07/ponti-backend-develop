package pkgsmtp

import (
	"fmt"
	"os"
)

// Bootstrap inicializa el servicio SMTP con la configuración necesaria.
func Bootstrap(host, port, from, username, password, identity string) (Service, error) {
	// Si no se proporcionan parámetros, se leen de variables de entorno
	if host == "" {
		host = os.Getenv("SMTP_HOST")
	}
	if port == "" {
		port = os.Getenv("SMTP_PORT")
	}
	if from == "" {
		from = os.Getenv("SMTP_FROM")
	}
	if username == "" {
		username = os.Getenv("SMTP_USERNAME")
	}
	if password == "" {
		password = os.Getenv("SMTP_PASSWORD")
	}
	if identity == "" {
		identity = os.Getenv("SMTP_IDENTITY")
	}

	// Construimos la config con lo que tenemos
	config := newConfig(
		host,
		port,
		from,
		username,
		password,
		identity,
	)

	// Validamos que la configuración sea correcta
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("SMTP config error: %w", err)
	}

	// Instanciamos el servicio usando la config
	return newService(config)
}
