package manager

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateManager(ctx context.Context, c *domain.Manager) (int64, error) {
	if c == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "manager is nil", nil)
	}
	model := models.FromDomain(c)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create manager", err)
	}
	return model.ID, nil
}

func (r *repository) ListManagers(ctx context.Context) ([]domain.Manager, error) {
	var list []models.Manager
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list customers", err)
	}
	result := make([]domain.Manager, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *repository) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	var model models.Manager
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("manager with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get manager", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateManager(ctx context.Context, c *domain.Manager) error {
	if c == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "manager is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Manager{}).
		Where("id = ?", c.ID).
		Updates(models.FromDomain(c))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update manager", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("manager with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *repository) DeleteManager(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Manager{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete manager", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("manager with id %d does not exist", id), nil)
	}
	return nil
}
