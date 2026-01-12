package pkgmysql

import "fmt"

type config struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

// newConfig crea una nueva instancia de configuración (actualmente no se usa directamente).
func newConfig() *config {
	return &config{}
}

// dsn construye el Data Source Name (DSN) para conectar a MySQL.
func (c config) dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

func (c config) Validate() error {
	if c.User == "" || c.Password == "" || c.Host == "" || c.Port == "" || c.Database == "" {
		return fmt.Errorf("incomplete MySQL configuration")
	}
	return nil
}
