package classtype

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
)

// RepositoryPort is the repository contract for ClassType.
type RepositoryPort interface {
	ListClassTypes(context.Context) ([]domain.ClassType, error)
	CreateClassType(context.Context, *domain.ClassType) (int64, error)
	UpdateClassType(context.Context, *domain.ClassType) error
	DeleteClassType(context.Context, int64) error
}
type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}
func (u *UseCases) ListClassTypes(ctx context.Context) ([]domain.ClassType, error) {
	return u.repo.ListClassTypes(ctx)
}
func (u *UseCases) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	return u.repo.CreateClassType(ctx, c)
}
func (u *UseCases) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	return u.repo.UpdateClassType(ctx, c)
}
func (u *UseCases) DeleteClassType(ctx context.Context, id int64) error {
	return u.repo.DeleteClassType(ctx, id)
}
