package pkgsmtp

import (
	"context"
	"net/smtp"
)

// Config define la interfaz que debe cumplir la configuración SMTP
type Config interface {
	GetSMTPServer() string
	GetAuth() smtp.Auth
	GetFrom() string
	GetPort() string
	Validate() error
}

type Service interface {
	SendEmail(context.Context, *Email) error
}
