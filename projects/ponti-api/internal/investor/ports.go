package investor

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

// UseCases defines business operations for Investor.
type UseCases interface {
	CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error)
	ListInvestors(ctx context.Context) ([]domain.Investor, error)
	GetInvestor(ctx context.Context, id int64) (*domain.Investor, error)
	UpdateInvestor(ctx context.Context, inv *domain.Investor) error
	DeleteInvestor(ctx context.Context, id int64) error
}

// Repository defines data persistence operations for Investor.
type Repository interface {
	CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error)
	ListInvestors(ctx context.Context) ([]domain.Investor, error)
	GetInvestor(ctx context.Context, id int64) (*domain.Investor, error)
	UpdateInvestor(ctx context.Context, inv *domain.Investor) error
	DeleteInvestor(ctx context.Context, id int64) error
}
