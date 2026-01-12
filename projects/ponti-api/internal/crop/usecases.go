package crop

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases creates a new instance of Crop use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateCrop(ctx context.Context, c *domain.Crop) (int64, error) {
	return u.repo.CreateCrop(ctx, c)
}

func (u *useCases) ListCrops(ctx context.Context) ([]domain.Crop, error) {
	return u.repo.ListCrops(ctx)
}

func (u *useCases) GetCrop(ctx context.Context, id int64) (*domain.Crop, error) {
	return u.repo.GetCrop(ctx, id)
}

func (u *useCases) UpdateCrop(ctx context.Context, c *domain.Crop) error {
	return u.repo.UpdateCrop(ctx, c)
}

func (u *useCases) DeleteCrop(ctx context.Context, id int64) error {
	return u.repo.DeleteCrop(ctx, id)
}
