package pkgpg

import (
	"fmt"
)

type Config interface {
	Validate() error
	DNS() string
}

type config struct {
	Host     string
	User     string
	Password string
	DbName   string
	Port     string
}

// newConfig crea una nueva configuración con los valores proporcionados
func newConfig(user, password, host, port, dbName string) Config {
	return &config{
		Host:     host,
		User:     user,
		Password: password,
		DbName:   dbName,
		Port:     port,
	}
}

// DNS genera la cadena de conexión para PostgreSQL
func (c *config) DNS() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DbName)
}

// Validate valida que los campos necesarios estén presentes
func (c *config) Validate() error {
	if c.User == "" {
		return fmt.Errorf("POSTGRES_USER environmente variable is empty")
	}
	if c.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD environmente variable is empty")
	}
	if c.Host == "" {
		return fmt.Errorf("POSTGRES_HOST environmente variable is empty")
	}
	if c.Port == "" {
		return fmt.Errorf("POSTGRES_PORT environmente variable is empty")
	}
	if c.DbName == "" {
		return fmt.Errorf("DbName is empty")
	}
	return nil
}
