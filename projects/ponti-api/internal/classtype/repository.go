package classtype

import (
	"context"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
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

func (r *Repository) ListClassTypes(ctx context.Context) ([]domain.ClassType, error) {
	var classTypes []models.ClassType
	if err := r.db.Client().WithContext(ctx).Find(&classTypes).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list class types", err)
	}
	res := make([]domain.ClassType, len(classTypes))
	for i := range classTypes {
		res[i] = *classTypes[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create class type", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.ClassType{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check class type existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("class type %d not found", c.ID), nil)
		}
		if err := tx.Model(&models.ClassType{}).
			Where("id = ?", c.ID).
			Updates(map[string]any{
				"name":       c.Name,
				"updated_by": c.UpdatedBy,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update class type", err)
		}
		return nil
	})
}

func (r *Repository) DeleteClassType(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.ClassType{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check class type existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("class type %d not found", id), nil)
		}
		if err := tx.Delete(&models.ClassType{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete class type", err)
		}
		return nil
	})
}
