package crop

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateCrop(ctx context.Context, c *domain.Crop) (int64, error) {
	if c == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "crop is nil", nil)
	}
	model := models.FromDomainCrop(c)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create crop", err)
	}
	return model.ID, nil
}

func (r *repository) ListCrops(ctx context.Context) ([]domain.Crop, error) {
	var list []models.Crop
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list crops", err)
	}
	result := make([]domain.Crop, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *repository) GetCrop(ctx context.Context, id int64) (*domain.Crop, error) {
	var model models.Crop
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("crop with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get crop", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateCrop(ctx context.Context, c *domain.Crop) error {
	if c == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "crop is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Crop{}).
		Where("id = ?", c.ID).
		Updates(models.FromDomainCrop(c))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update crop", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("crop with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *repository) DeleteCrop(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Crop{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete crop", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("crop with id %d does not exist", id), nil)
	}
	return nil
}
