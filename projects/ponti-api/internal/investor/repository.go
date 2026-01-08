package investor

import (
	"context"
	"errors"
	"fmt"

	gorm "gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	if inv == nil {
		return 0, types.NewError(types.ErrValidation, "investor is nil", nil)
	}
	model := models.FromDomain(inv)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create investor", err)
	}
	return model.ID, nil
}

// En internal/investor/Repository.go
func (r *Repository) ListInvestors(ctx context.Context) ([]domain.ListedInvestor, error) {
	var list []models.Investor
	if err := r.db.Client().
		WithContext(ctx).
		Model(&models.Investor{}).
		Select("id, name").
		Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list investors", err)
	}

	result := make([]domain.ListedInvestor, len(list))
	for i, m := range list {
		result[i] = domain.ListedInvestor{
			ID:   m.ID,
			Name: m.Name,
		}
	}
	return result, nil
}

func (r *Repository) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	var model models.Investor
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("investor with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get investor", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	if inv == nil {
		return types.NewError(types.ErrValidation, "investor is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Investor{}).
		Where("id = ?", inv.ID).
		Updates(models.FromDomain(inv))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update investor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("investor with id %d does not exist", inv.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteInvestor(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Investor{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete investor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("investor with id %d does not exist", id), nil)
	}
	return nil
}
