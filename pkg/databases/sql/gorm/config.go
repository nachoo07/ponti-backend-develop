package pkggorm

import (
	"fmt"
)

// DBType define los tipos de bases de datos soportadas
type DBType string

const (
	Postgres DBType = "postgres"
	MySQL    DBType = "mysql"
	SQLite   DBType = "sqlite"
)

// Config es una implementación concreta de Config
type Config struct {
	dbType     DBType
	sslMode    string
	host       string
	user       string
	password   string
	dbname     string
	port       int
	sqlitePath string
}

// newConfig crea una nueva instancia de Config
func newConfig(dbType DBType, host, user, password, dbname string, port int, sqlitePath, sslMode string) *Config {
	return &Config{
		dbType:     dbType,
		sslMode:    sslMode,
		host:       host,
		user:       user,
		password:   password,
		dbname:     dbname,
		port:       port,
		sqlitePath: sqlitePath,
	}
}

// Métodos de `Config` para implementar la interfaz Config
func (c *Config) GetDBType() DBType {
	return c.dbType
}

func (c *Config) GetHost() string {
	return c.host
}

func (c *Config) GetUser() string {
	return c.user
}

func (c *Config) GetSSLMode() string {
	return c.sslMode
}

func (c *Config) GetPassword() string {
	return c.password
}

func (c *Config) GetDBName() string {
	return c.dbname
}

func (c *Config) GetPort() int {
	return c.port
}

func (c *Config) GetSQLitePath() string {
	return c.sqlitePath
}

// Validate verifica si la Configuración es válida
func (c *Config) Validate() error {
	/*if os.Getenv("K_SERVICE") != "" {
		if c.user == "" || c.dbname == "" || os.Getenv("INSTANCE_CONNECTION_NAME") == "" {
			return fmt.Errorf("incomplete %s Configuration", c.dbType)
		}
		return nil
	}*/

	switch c.dbType {
	case Postgres, MySQL:
		if c.host == "" || c.user == "" || c.password == "" || c.dbname == "" || c.port == 0 {
			return fmt.Errorf("incomplete %s Configuration", c.dbType)
		}
	case SQLite:
		if c.sqlitePath == "" {
			return fmt.Errorf("sqlite path is required")
		}
	default:
		return fmt.Errorf("unsupported database type: %s", c.dbType)
	}
	return nil
}
