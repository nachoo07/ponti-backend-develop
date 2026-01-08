package leasetype

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
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

func (r *Repository) CreateLeaseType(ctx context.Context, lt *domain.LeaseType) (int64, error) {
	if lt == nil {
		return 0, types.NewError(types.ErrValidation, "lease type is nil", nil)
	}
	model := models.FromDomainLeaseType(lt)
	model.CreatedBy = lt.CreatedBy
	model.UpdatedBy = lt.UpdatedBy

	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create lease type", err)
	}
	return model.ID, nil
}

func (r *Repository) ListLeaseTypes(ctx context.Context) ([]domain.LeaseType, error) {
	var list []models.LeaseType
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lease types", err)
	}
	result := make([]domain.LeaseType, 0, len(list))
	for _, lt := range list {
		result = append(result, *lt.ToDomain())
	}
	return result, nil
}

func (r *Repository) GetLeaseType(ctx context.Context, id int64) (*domain.LeaseType, error) {
	var model models.LeaseType
	err := r.db.Client().WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("lease type with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get lease type", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateLeaseType(ctx context.Context, lt *domain.LeaseType) error {
	if lt == nil {
		return types.NewError(types.ErrValidation, "lease type is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.LeaseType{}).
		Where("id = ?", lt.ID).
		Updates(models.FromDomainLeaseType(lt))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update lease type", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("lease type with id %d does not exist", lt.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteLeaseType(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.LeaseType{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete lease type", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("lease type with id %d does not exist", id), nil)
	}
	return nil
}
