package dollar

import (
	"context"
	"errors"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
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

func (r *Repository) ListByProject(ctx context.Context, projectID int64) ([]domain.DollarAverage, error) {
	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.ProjectDollarValue{}).
		Where("project_id = ?", projectID)

	var rows []models.ProjectDollarValue
	if err := tx.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list project dollar values", err)
	}

	out := make([]domain.DollarAverage, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.DollarAverage) (int64, error) {
	m := models.FromDomain(item)
	if err := r.db.Client().WithContext(ctx).Create(m).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create dollar value", err)
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.DollarAverage) error {
	if item.ID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid id", nil)
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.ProjectDollarValue{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "project dollar value not found", nil)
		}

		// Map ONLY the updatable fields (GORM will update Base automatically)
		if err := tx.Model(&models.ProjectDollarValue{}).
			Where("id = ?", item.ID).
			Updates(map[string]any{
				"project_id":    item.ProjectID,
				"year":          item.Year,
				"month":         item.Month,
				"start_value":   item.StartValue,
				"end_value":     item.EndValue,
				"average_value": item.AvgValue,
				"updated_at":    time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update project dollar value", err)
		}
		return nil
	})
}

func (r *Repository) GetByComposite(ctx context.Context, projectID, year int64, month string) (*domain.DollarAverage, error) {
	var m models.ProjectDollarValue
	err := r.db.Client().WithContext(ctx).
		Where("project_id = ? AND year = ? AND month = ?", projectID, year, month).
		First(&m).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to query by composite key", err)
	}
	return m.ToDomain(), nil
}
