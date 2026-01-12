package pkgjwt

import (
	"fmt"
	"os"
	"strconv"
)

func Bootstrap(secret string, accessExpirationMinutes, refreshExpirationMinutes int) (Service, error) {
	if secret == "" {
		secret = os.Getenv("JWT_SECRET_KEY")
	}
	if accessExpirationMinutes == 0 {
		accessExpirationMinutes, _ = strconv.Atoi(os.Getenv("JWT_DEFAULT_ACCESS_EXPIRATION_MINUTES"))
	}
	if refreshExpirationMinutes == 0 {
		refreshExpirationMinutes, _ = strconv.Atoi(os.Getenv("JWT_DEFAULT_REFRESH_EXPIRATION_MINUTES"))
	}

	config := newConfig(
		secret,
		accessExpirationMinutes,
		refreshExpirationMinutes,
	)

	// Validar la configuración
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid JWT configuration: %w", err)
	}

	// Crear el servicio JWT
	return newService(config)
}
