package investor

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases creates a new instance of Investor use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	return u.repo.CreateInvestor(ctx, inv)
}

func (u *useCases) ListInvestors(ctx context.Context) ([]domain.Investor, error) {
	return u.repo.ListInvestors(ctx)
}

func (u *useCases) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	return u.repo.GetInvestor(ctx, id)
}

func (u *useCases) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	return u.repo.UpdateInvestor(ctx, inv)
}

func (u *useCases) DeleteInvestor(ctx context.Context, id int64) error {
	return u.repo.DeleteInvestor(ctx, id)
}
