package pkgsmtp

import (
	"fmt"
	"net/smtp"
)

// config implementa la interfaz Config
type config struct {
	smtpServer string
	auth       smtp.Auth
	from       string
	port       string
}

// newConfig crea una implementación de config.
func newConfig(smtpServer, port, from, username, password, identity string) Config {
	auth := smtp.PlainAuth(identity, username, password, smtpServer)

	return &config{
		smtpServer: smtpServer,
		auth:       auth,
		from:       from,
		port:       port,
	}
}

// GetSMTPServer devuelve la dirección del servidor SMTP
func (c *config) GetSMTPServer() string {
	return c.smtpServer
}

// GetAuth devuelve la autenticación SMTP configurada
func (c *config) GetAuth() smtp.Auth {
	return c.auth
}

// GetFrom devuelve la dirección de correo del remitente
func (c *config) GetFrom() string {
	return c.from
}

// GetPort devuelve el puerto para conectarse al servidor SMTP
func (c *config) GetPort() string {
	return c.port
}

// Validate verifica que la configuración sea válida
func (c *config) Validate() error {
	if c.smtpServer == "" {
		return fmt.Errorf("SMTP server is not configured")
	}
	if c.auth == nil {
		return fmt.Errorf("SMTP auth is not configured")
	}
	if c.from == "" {
		return fmt.Errorf("SMTP from address is not configured")
	}
	if c.port == "" {
		return fmt.Errorf("SMTP port is not configured")
	}
	return nil
}
