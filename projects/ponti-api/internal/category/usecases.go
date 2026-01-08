package category

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/usecases/domain"
)

// RepositoryPort is the repository contract for Category.
type RepositoryPort interface {
	ListCategories(context.Context) ([]domain.Category, error)
	CreateCategory(context.Context, *domain.Category) (int64, error)
	UpdateCategory(context.Context, *domain.Category) error
	DeleteCategory(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return u.repo.ListCategories(ctx)
}
func (u *UseCases) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	return u.repo.CreateCategory(ctx, c)
}
func (u *UseCases) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return u.repo.UpdateCategory(ctx, c)
}
func (u *UseCases) DeleteCategory(ctx context.Context, id int64) error {
	return u.repo.DeleteCategory(ctx, id)
}
