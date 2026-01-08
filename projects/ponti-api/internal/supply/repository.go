package supply

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
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

// --- CREATE ---
func (r *Repository) CreateSupply(ctx context.Context, s *domain.Supply) (int64, error) {
	var id int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := models.FromDomain(s)
		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create supply", err)
		}
		id = model.ID
		return nil
	})
	return id, err
}

func (r *Repository) CreateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	userID, err := sharedmodels.ConvertStringToID(ctx)
	if err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		modelsSlice := make([]*models.Supply, len(supplies))
		for i := range supplies {
			modelsSlice[i] = models.FromDomain(&supplies[i])
			modelsSlice[i].CreatedBy = &userID
		}
		if err := tx.Create(modelsSlice).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to bulk create supplies", err)
		}
		return nil
	})
}

// --- GET ---
func (r *Repository) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	var m models.Supply
	if err := r.db.Client().WithContext(ctx).
		Preload("Category").
		Preload("Type").
		First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get supply", err)
	}
	return m.ToDomain(), nil
}

// --- UPDATE ---
func (r *Repository) UpdateSupply(ctx context.Context, s *domain.Supply) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", s.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", s.ID), nil)
		}
		updates := map[string]any{
			"name":        s.Name,
			"unit_id":     int64(s.UnitID),
			"price":       s.Price,
			"category_id": int64(s.CategoryID),
			"type_id":     s.Type.ID,
			"project_id":  s.ProjectID,
			"updated_by":  s.UpdatedBy,
		}
		if err := tx.Model(&models.Supply{}).
			Where("id = ?", s.ID).
			Updates(updates).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update supply", err)
		}
		return nil
	})
}

// --- DELETE ---
func (r *Repository) GetWorkordersBySupplyID(ctx context.Context, supplyID int64) (int64, error) {
	var count int64
	if err := r.db.Client().WithContext(ctx).
		Model(&workordermodels.Workorder{}).
		Joins("JOIN workorder_items ON workorder_items.workorder_id = workorders.id").
		Where("workorder_items.supply_id = ?", supplyID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return count, nil
}

func (r *Repository) DeleteSupply(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), nil)
		}
		if err := tx.Delete(&models.Supply{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete supply", err)
		}
		return nil
	})
}

// --- LIST CENTRALIZADO, con filtros y paginación ---
func (r *Repository) ListSuppliesPaginated(
	ctx context.Context,
	projectID, campaignID int64,
	mode string,
	page, perPage int,
) ([]domain.Supply, int64, error) {
	var supplies []models.Supply
	var total int64

	db := r.db.Client().WithContext(ctx).Model(&models.Supply{}).
		Preload("Category").
		Preload("Type")

	// Filtrado flexible
	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}

	// Total para paginación
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count supplies", err)
	}

	offset := (page - 1) * perPage
	if err := db.Offset(offset).Limit(perPage).Order("name").Find(&supplies).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list supplies with filters", err)
	}

	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}
	return res, total, nil
}

func (r *Repository) UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range supplies {
			updates := map[string]any{
				"name":        supplies[i].Name,
				"unit_id":     int64(supplies[i].UnitID),
				"price":       supplies[i].Price,
				"category_id": int64(supplies[i].CategoryID),
				"type_id":     supplies[i].Type.ID,
				"project_id":  supplies[i].ProjectID,
				"updated_by":  supplies[i].UpdatedBy,
			}
			res := tx.Model(&models.Supply{}).
				Where("id = ?", supplies[i].ID).
				Updates(updates)
			if res.Error != nil {
				return types.NewError(types.ErrInternal, fmt.Sprintf("failed to update supply id %d", supplies[i].ID), res.Error)
			}
			if res.RowsAffected == 0 {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", supplies[i].ID), nil)
			}
		}
		return nil
	})
}

func (r *Repository) ListAllSupplies(ctx context.Context, projectID int64) ([]domain.Supply, int64, error) {
	base := r.db.Client().WithContext(ctx).Model(&models.Supply{})

	if projectID > 0 {
		base = base.Where("project_id = ?", projectID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count supplies", err)
	}

	var rows []models.Supply
	db := base.
		Preload("Category").
		Preload("Type").
		Order("name")

	if err := db.Find(&rows).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list all supplies", err)
	}

	out := make([]domain.Supply, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out, total, nil
}
