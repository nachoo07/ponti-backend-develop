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

// Config es la interfaz para manejar configuraciones del cliente GORM
type Config interface {
	GetDBType() DBType
	GetHost() string
	GetUser() string
	GetPassword() string
	GetDBName() string
	GetPort() int
	GetSQLitePath() string
	Validate() error
}

// config es una implementación concreta de Config
type config struct {
	dbType     DBType
	host       string
	user       string
	password   string
	dbname     string
	port       int
	sqlitePath string
}

// newConfig crea una nueva instancia de Config
func newConfig(dbType DBType, host, user, password, dbname string, port int, sqlitePath string) Config {
	return &config{
		dbType:     dbType,
		host:       host,
		user:       user,
		password:   password,
		dbname:     dbname,
		port:       port,
		sqlitePath: sqlitePath,
	}
}

// Métodos de `config` para implementar la interfaz Config
func (c *config) GetDBType() DBType {
	return c.dbType
}

func (c *config) GetHost() string {
	return c.host
}

func (c *config) GetUser() string {
	return c.user
}

func (c *config) GetPassword() string {
	return c.password
}

func (c *config) GetDBName() string {
	return c.dbname
}

func (c *config) GetPort() int {
	return c.port
}

func (c *config) GetSQLitePath() string {
	return c.sqlitePath
}

// Validate verifica si la configuración es válida
func (c *config) Validate() error {
	switch c.dbType {
	case Postgres, MySQL:
		if c.host == "" || c.user == "" || c.password == "" || c.dbname == "" || c.port == 0 {
			return fmt.Errorf("incomplete %s configuration", c.dbType)
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
