package investor

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type RepositoryPort interface {
	CreateInvestor(context.Context, *domain.Investor) (int64, error)
	ListInvestors(context.Context) ([]domain.ListedInvestor, error)
	GetInvestor(context.Context, int64) (*domain.Investor, error)
	UpdateInvestor(context.Context, *domain.Investor) error
	DeleteInvestor(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	return u.repo.CreateInvestor(ctx, inv)
}

func (u *UseCases) ListInvestors(ctx context.Context) ([]domain.ListedInvestor, error) {
	return u.repo.ListInvestors(ctx)
}

func (u *UseCases) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	return u.repo.GetInvestor(ctx, id)
}

func (u *UseCases) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	return u.repo.UpdateInvestor(ctx, inv)
}

func (u *UseCases) DeleteInvestor(ctx context.Context, id int64) error {
	return u.repo.DeleteInvestor(ctx, id)
}
