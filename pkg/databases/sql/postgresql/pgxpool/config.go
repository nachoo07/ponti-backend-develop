package pkgpostgresql

import (
	"fmt"
	"log"
)



type config struct {
	Host          string
	User          string
	Password      string
	DbName        string
	Port          string
	MigrationsDir string
}

// newConfig crea una nueva configuración con los valores proporcionados
func newConfig(user, password, host, port, migrationsDir, dbName string) Config {
	return &config{
		Host:          host,
		User:          user,
		Password:      password,
		DbName:        dbName,
		Port:          port,
		MigrationsDir: migrationsDir,
	}
}

// DNS genera la cadena de conexión para PostgreSQL
func (c *config) DNS() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.DbName)
}

func (c *config) GetHost() string {
	return c.Host
}

func (c *config) GetUser() string {
	return c.User
}

func (c *config) GetPassword() string {
	return c.Password
}

func (c *config) GetDbName() string {
	return c.DbName
}

func (c *config) GetPort() string {
	return c.Port
}

func (c *config) GetMigrationsDir() string {
	return c.MigrationsDir
}

// Validate valida que los campos necesarios estén presentes
func (c *config) Validate() error {
	if c.User == "" {
		return fmt.Errorf("POSTGRES_USER environment variable is empty")
	}
	if c.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD environment variable is empty")
	}
	if c.Host == "" {
		return fmt.Errorf("POSTGRES_HOST environment variable is empty")
	}
	if c.Port == "" {
		return fmt.Errorf("POSTGRES_PORT environment variable is empty")
	}
	if c.DbName == "" {
		return fmt.Errorf("POSTGRES_DB environment variable is empty")
	}
	if c.MigrationsDir == "" {
		log.Println("Warning: MIGRATIONS_DIR environment variable is empty")
	}
	return nil
}
