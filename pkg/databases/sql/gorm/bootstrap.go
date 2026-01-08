package pkggorm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Bootstrap inicializa la base de datos sin aplicar migraciones automáticamente.
func Bootstrap(dbTypeStr, host, user, password, name, sslMode string, port int) (*Repository, error) {
	if dbTypeStr == "" {
		dbTypeStr = strings.ToLower(os.Getenv("DB_TYPE"))
	}

	var dbType DBType
	switch dbTypeStr {
	case "postgres":
		dbType = Postgres
	case "mysql":
		dbType = MySQL
	case "sqlite":
		dbType = SQLite
	default:
		return nil, fmt.Errorf("unsupported DB_TYPE: %s", dbTypeStr)
	}

	var config *Config
	switch dbType {
	case Postgres, MySQL:
		if host == "" {
			host = os.Getenv("DB_HOST")
		}
		if sslMode == "" {
			sslMode = os.Getenv("SSL_MODE")
		}
		if user == "" {
			user = os.Getenv("DB_USER")
		}
		if password == "" {
			password = os.Getenv("DB_PASSWORD")
		}
		if name == "" {
			name = os.Getenv("DB_NAME")
		}
		if port == 0 {
			port, _ = strconv.Atoi(os.Getenv("DB_PORT"))
		}

		config = newConfig(
			dbType,
			host,
			user,
			password,
			name,
			port,
			"",
			sslMode,
		)
	case SQLite:
		config = newConfig(
			dbType,
			"",
			"",
			"",
			"",
			0,
			os.Getenv("SQLITE_PATH"),
			"",
		)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newRepository(config)
}
