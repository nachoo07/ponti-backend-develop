package field

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
)

type RepositoryPort interface {
	CreateField(ctx context.Context, f *domain.Field) (int64, error)
	ListFields(ctx context.Context) ([]domain.Field, error)
	GetField(ctx context.Context, id int64) (*domain.Field, error)
	UpdateField(ctx context.Context, f *domain.Field) error
	DeleteField(ctx context.Context, id int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	return u.repo.CreateField(ctx, f)
}
func (u *UseCases) ListFields(ctx context.Context) ([]domain.Field, error) {
	return u.repo.ListFields(ctx)
}
func (u *UseCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	return u.repo.GetField(ctx, id)
}
func (u *UseCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}
func (u *UseCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}
