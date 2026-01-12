package pkgpostgresql

import "os"

func Bootstrap(user, password, host, port, migrationsDir, dbName string) (Repository, error) {
	// Si algún parámetro es vacío, se usa os.Getenv para obtener el valor
	if user == "" {
		user = os.Getenv("POSTGRES_USER")
	}
	if password == "" {
		password = os.Getenv("POSTGRES_PASSWORD")
	}
	if host == "" {
		host = os.Getenv("POSTGRES_HOST")
	}
	if port == "" {
		port = os.Getenv("POSTGRES_PORT")
	}
	if migrationsDir == "" {
		migrationsDir = os.Getenv("POSTGRES_MIGRATIONS_DIR")
	}
	if dbName == "" {
		dbName = os.Getenv("POSTGRES_DB")
	}

	// Crear la configuración
	config := newConfig(
		user,
		password,
		host,
		port,
		migrationsDir,
		dbName,
	)

	// Validar la configuración
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Retornar el repositorio con la configuración validada
	return newRepository(config)
}
