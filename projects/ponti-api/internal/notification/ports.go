package notification

import (
	"context"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/notification/usecases/domain"
)

type SmtpService interface {
	SendEmail(context.Context, *domain.Email) error
}

type UseCases interface {
	SendEmail(context.Context, string, string, string) error
}

type Cache interface {
	StoreRefreshToken(context.Context, string, string, time.Time) error
	RetrieveRefreshToken(context.Context, string) (string, error)
	Close()
}
