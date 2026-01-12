package manager

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

// UseCases define las operaciones de negocio para Manager.
type UseCases interface {
	CreateManager(ctx context.Context, c *domain.Manager) (int64, error)
	ListManagers(ctx context.Context) ([]domain.Manager, error)
	GetManager(ctx context.Context, id int64) (*domain.Manager, error)
	UpdateManager(ctx context.Context, c *domain.Manager) error
	DeleteManager(ctx context.Context, id int64) error
}

// Repository define las operaciones para Manager.
type Repository interface {
	CreateManager(ctx context.Context, c *domain.Manager) (int64, error)
	ListManagers(ctx context.Context) ([]domain.Manager, error)
	GetManager(ctx context.Context, id int64) (*domain.Manager, error)
	UpdateManager(ctx context.Context, c *domain.Manager) error
	DeleteManager(ctx context.Context, id int64) error
}
