package leasetype

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
)

type RepositoryPort interface {
	CreateLeaseType(context.Context, *domain.LeaseType) (int64, error)
	ListLeaseTypes(context.Context) ([]domain.LeaseType, error)
	GetLeaseType(context.Context, int64) (*domain.LeaseType, error)
	UpdateLeaseType(context.Context, *domain.LeaseType) error
	DeleteLeaseType(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateLeaseType(ctx context.Context, lt *domain.LeaseType) (int64, error) {
	return u.repo.CreateLeaseType(ctx, lt)
}
func (u *UseCases) ListLeaseTypes(ctx context.Context) ([]domain.LeaseType, error) {
	return u.repo.ListLeaseTypes(ctx)
}
func (u *UseCases) GetLeaseType(ctx context.Context, id int64) (*domain.LeaseType, error) {
	return u.repo.GetLeaseType(ctx, id)
}
func (u *UseCases) UpdateLeaseType(ctx context.Context, lt *domain.LeaseType) error {
	return u.repo.UpdateLeaseType(ctx, lt)
}
func (u *UseCases) DeleteLeaseType(ctx context.Context, id int64) error {
	return u.repo.DeleteLeaseType(ctx, id)
}
