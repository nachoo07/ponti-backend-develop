package pkggorm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Bootstrap inicializa la base de datos sin aplicar migraciones automáticamente.
func Bootstrap(dbTypeStr, name, password, user, host string, port int) (Repository, error) {
	if dbTypeStr == "" {
		dbTypeStr = strings.ToLower(os.Getenv("GORM_TYPE"))
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

	var config Config
	switch dbType {
	case Postgres, MySQL:

		if host == "" {
			host = os.Getenv("GORM_HOST")
		}

		if user == "" {
			user = os.Getenv("GORM_USER")
		}
		if password == "" {
			password = os.Getenv("GORM_PASSWORD")
		}
		if name == "" {
			name = os.Getenv("GORM_NAME")
		}
		if port == 0 {
			port, _ = strconv.Atoi(os.Getenv("GORM_PORT"))
		}

		config = newConfig(
			dbType,
			host,
			user,
			password,
			name,
			port,
			"",
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
		)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	repo, err := newRepository(config)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
