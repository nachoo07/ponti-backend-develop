package manager

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases crea una instancia de los casos de uso para Manager.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateManager(ctx context.Context, c *domain.Manager) (int64, error) {
	return u.repo.CreateManager(ctx, c)
}

func (u *useCases) ListManagers(ctx context.Context) ([]domain.Manager, error) {
	return u.repo.ListManagers(ctx)
}

func (u *useCases) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	return u.repo.GetManager(ctx, id)
}

func (u *useCases) UpdateManager(ctx context.Context, c *domain.Manager) error {
	return u.repo.UpdateManager(ctx, c)
}

func (u *useCases) DeleteManager(ctx context.Context, id int64) error {
	return u.repo.DeleteManager(ctx, id)
}
