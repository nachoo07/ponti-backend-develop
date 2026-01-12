package crop

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

type UseCases interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context) ([]domain.Crop, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	DeleteCrop(context.Context, int64) error
}

type Repository interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context) ([]domain.Crop, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	DeleteCrop(context.Context, int64) error
}
