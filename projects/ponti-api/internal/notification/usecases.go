package notification

import (
	"context"
	"fmt"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/notification/usecases/domain"
)

type useCases struct {
	smtpService SmtpService
}

// NewUseCases inicializa y retorna una implementación de la interfaz useCases.
func NewUseCases(ss SmtpService) UseCases {
	return &useCases{
		smtpService: ss,
	}
}

func (u *useCases) SendEmail(ctx context.Context, address, subject, body string) error {
	email := &domain.Email{
		Address: address,
		Subject: subject,
		Body:    body,
	}

	if err := u.smtpService.SendEmail(ctx, email); err != nil {
		return fmt.Errorf("failed to send test link email: %w", err)
	}

	return nil
}
