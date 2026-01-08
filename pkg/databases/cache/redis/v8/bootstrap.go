package pkgredis

import (
	"os"
	"strconv"
)

func Bootstrap(address, password string, dbName int) (Cache, error) {
	// 1. Intentamos leer la variable de entorno si no viene por parámetro
	if address == "" {
		address = os.Getenv("REDIS_ADDRESS")
	}

	// 🚨 EL ARREGLO ESTÁ AQUÍ 🚨
	// Si después de intentar leer la variable, la dirección sigue vacía,
	// significa que en este entorno NO queremos usar Redis.
	// Devolvemos 'nil' (sin caché) y 'nil' (sin error).
	if address == "" {
		return nil, nil
	}
	// ---------------------------

	if password == "" {
		password = os.Getenv("REDIS_PASSWORD")
	}
	if dbName == 0 {
		dbName, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
	}

	config := newConfig(
		address,
		password,
		dbName,
	)

	// Validamos la configuración solo si realmente vamos a usar Redis
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return NewCache(config)
}