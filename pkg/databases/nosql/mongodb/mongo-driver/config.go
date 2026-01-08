package pkgmongo

import (
	"fmt"
	"net/url"
)

type config struct {
	User         string
	Password     string
	Host         string
	Port         string
	DatabaseName string
}

func newConfig(user, password, host, port, databaseName string) *config {
	return &config{
		User:         user,
		Password:     password,
		Host:         host,
		Port:         port,
		DatabaseName: databaseName,
	}
}

func (c *config) GetUser() string {
	return c.User
}

func (c *config) GetPassword() string {
	return c.Password
}

func (c *config) GetHost() string {
	return c.Host
}

func (c *config) GetPort() string {
	return c.Port
}

func (c *config) GetDatabaseName() string {
	return c.DatabaseName
}

// DSN retorna la URL de conexión a MongoDB con parámetros adicionales.
func (c *config) DSN() string {
	u := &url.URL{
		Scheme: "mongodb",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.DatabaseName,
	}

	// Parámetros de consulta
	q := u.Query()
	// Especifica el authSource para la autenticación
	q.Set("authSource", c.DatabaseName)
	// Otras opciones útiles
	q.Set("retryWrites", "true")
	q.Set("w", "majority")
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *config) Database() string {
	return c.DatabaseName
}

func (c *config) Validate() error {
	if c.User == "" || c.Password == "" || c.Host == "" || c.Port == "" || c.DatabaseName == "" {
		return fmt.Errorf("incomplete MongoDB configuration")
	}
	return nil
}
