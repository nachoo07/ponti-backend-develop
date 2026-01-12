package pkgsmtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"sync"
)

var (
	instance Service
	once     sync.Once
	initErr  error
)

// service representa el servicio SMTP que envía correos.
type service struct {
	config Config
}

// newService crea una nueva instancia del servicio SMTP usando la configuración proporcionada.
func newService(config Config) (Service, error) {
	once.Do(func() {
		instance = &service{
			config: config,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

// SendEmail envía un correo electrónico usando el contenido de data (To, Subject, Body, etc.).
func (s *service) SendEmail(ctx context.Context, data *Email) error {
	// Construir el mensaje en formato RFC822
	// Nota: Aquí puedes agregar cabeceras MIME adicionales si quieres enviar HTML, etc.
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		data.Address,
		data.Subject,
		data.Body,
	))

	// Obtener información de config
	host := s.config.GetSMTPServer()
	port := s.config.GetPort()
	auth := s.config.GetAuth()
	from := s.config.GetFrom()

	// Determinar entorno (desarrollo / producción) según variable de entorno STAGE
	stage := os.Getenv("APP_ENV")
	if stage == "dev" {
		// Modo Desarrollo: Sin TLS
		client, err := smtp.Dial(fmt.Sprintf("%s:%s", host, port))
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer client.Quit()

		// Autenticación solo si no es mailhog / localhost
		if host != "mailhog" && host != "localhost" {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("failed to authenticate with SMTP server: %w", err)
			}
		}

		// Configurar el remitente
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}
		// Configurar el destinatario
		if err := client.Rcpt(data.Address); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		// Escribir el mensaje en el servidor SMTP
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get SMTP data writer: %w", err)
		}
		if _, err := w.Write(msg); err != nil {
			return fmt.Errorf("failed to write email message: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("failed to close email message writer: %w", err)
		}

		fmt.Printf("Email sent to %s (Development Mode)\n", data.Address)
		return nil
	}

	// Modo Producción: Usar TLS
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", host, port), &tls.Config{
		InsecureSkipVerify: true, // En producción, usa certificados válidos y pon esto en false.
	})
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Crear cliente SMTP sobre TLS
	client, err := smtp.NewClient(conn, s.config.GetSMTPServer())
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Autenticarse
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate with SMTP server: %w", err)
	}

	// Remitente y destinatario
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := client.Rcpt(data.Address); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Escribir el mensaje
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get SMTP data writer: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write email message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close email message writer: %w", err)
	}

	fmt.Printf("Email sent to %s\n", data.Address)
	return nil
}
