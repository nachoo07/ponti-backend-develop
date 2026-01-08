package crop

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

type RepositoryPort interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context) ([]domain.Crop, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	DeleteCrop(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases creates a new instance of Crop use cases.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateCrop(ctx context.Context, c *domain.Crop) (int64, error) {
	return u.repo.CreateCrop(ctx, c)
}

func (u *UseCases) ListCrops(ctx context.Context) ([]domain.Crop, error) {
	return u.repo.ListCrops(ctx)
}

func (u *UseCases) GetCrop(ctx context.Context, id int64) (*domain.Crop, error) {
	return u.repo.GetCrop(ctx, id)
}

func (u *UseCases) UpdateCrop(ctx context.Context, c *domain.Crop) error {
	return u.repo.UpdateCrop(ctx, c)
}

func (u *UseCases) DeleteCrop(ctx context.Context, id int64) error {
	return u.repo.DeleteCrop(ctx, id)
}
