package crop

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCrop(ctx context.Context, c *domain.Crop) (int64, error) {
	if c == nil {
		return 0, types.NewError(types.ErrValidation, "crop is nil", nil)
	}
	model := models.FromDomainCrop(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create crop", err)
	}
	return model.ID, nil
}

func (r *Repository) ListCrops(ctx context.Context) ([]domain.Crop, error) {
	var list []models.Crop
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list crops", err)
	}
	result := make([]domain.Crop, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *Repository) GetCrop(ctx context.Context, id int64) (*domain.Crop, error) {
	var model models.Crop
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("crop with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get crop", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCrop(ctx context.Context, c *domain.Crop) error {
	if c == nil {
		return types.NewError(types.ErrValidation, "crop is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Crop{}).
		Where("id = ?", c.ID).
		Updates(models.FromDomainCrop(c))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update crop", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("crop with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteCrop(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Crop{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete crop", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("crop with id %d does not exist", id), nil)
	}
	return nil
}
