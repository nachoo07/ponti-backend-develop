package field

import (
	"context"
	"errors"
	"fmt"
	"time"

	gorm "gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lotmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
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

// --- CREATE ---
func (r *Repository) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	var fieldID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := models.FromDomain(f)
		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create field", err)
		}
		fieldID = model.ID
		// Crear lots si vienen anidados
		for _, lot := range f.Lots {
			lotModel := lotmod.Lot{
				Name:           lot.Name,
				FieldID:        fieldID,
				Hectares:       lot.Hectares,
				PreviousCropID: lot.PreviousCrop.ID,
				CurrentCropID:  lot.CurrentCrop.ID,
				Season:         lot.Season,
			}
			if err := tx.Create(&lotModel).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create lot", err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return fieldID, nil
}

// --- LIST ---
func (r *Repository) ListFields(ctx context.Context) ([]domain.Field, error) {
	var list []models.Field
	if err := r.db.Client().WithContext(ctx).
		Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list fields", err)
	}
	result := make([]domain.Field, len(list))
	for i := range list {
		result[i] = *list[i].ToDomain()
	}
	return result, nil
}

// --- GET ---
func (r *Repository) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	var model models.Field
	err := r.db.Client().WithContext(ctx).
		First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("field with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get field", err)
	}
	return model.ToDomain(), nil
}

// --- UPDATE ---
func (r *Repository) UpdateField(ctx context.Context, f *domain.Field) error {
	if f == nil || f.ID <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid field id: %d", f.ID), nil)
	}
	model := models.FromDomain(f)
	model.ID = f.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Field{}).Where("id = ?", f.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check field existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("field %d not found", f.ID), nil)
		}
		if err := tx.Model(&models.Field{}).
			Where("id = ?", f.ID).
			Updates(map[string]any{
				"name":          f.Name,
				"lease_type_id": f.LeaseType.ID,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update field", err)
		}
		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteField(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid field id: %d", id), nil)
	}

	deletedBy, err := sharedmodels.ConvertStringToID(ctx)
	if err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Field{}).Where("id = ?", id).Updates(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete field", err)
		}
		return nil
	})
}
