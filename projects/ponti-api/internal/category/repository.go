// File: ./repository.go

package category

import (
	"context"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"gorm.io/gorm"
)

// GormEnginePort exposes the required DB interface.
type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []models.Category
	if err := r.db.Client().WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list categories", err)
	}
	res := make([]domain.Category, len(categories))
	for i := range categories {
		res[i] = *categories[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	model := models.FromDomain(c)
	// Se asegura de setear CreatedBy y UpdatedBy (otros campos los setea GORM)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create category", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Category{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check category existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("category %d not found", c.ID), nil)
		}
		// Solo actualiza el nombre (puedes extender para UpdatedBy, etc. si lo necesitas)
		if err := tx.Model(&models.Category{}).
			Where("id = ?", c.ID).
			Updates(map[string]any{
				"name":       c.Name,
				"updated_by": c.UpdatedBy,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update category", err)
		}
		return nil
	})
}

func (r *Repository) DeleteCategory(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Category{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check category existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("category %d not found", id), nil)
		}
		if err := tx.Delete(&models.Category{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete category", err)
		}
		return nil
	})
}
