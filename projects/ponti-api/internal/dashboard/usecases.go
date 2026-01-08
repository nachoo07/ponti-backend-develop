package dashboard

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type RepositoryPort interface {
	GetDashboard(context.Context, domain.DashboardFilter) (*domain.DashboardData, error)
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases creates a new instance of Dashboard use cases.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetDashboard(ctx context.Context, filter domain.DashboardFilter) (*domain.DashboardData, error) {
	return u.repo.GetDashboard(ctx, filter)
}
