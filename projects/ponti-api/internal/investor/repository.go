package investor

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository creates a new Investor repository instance.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	if inv == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "investor is nil", nil)
	}
	model := models.FromDomain(inv)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create investor", err)
	}
	return model.ID, nil
}

func (r *repository) ListInvestors(ctx context.Context) ([]domain.Investor, error) {
	var list []models.Investor
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list investors", err)
	}
	result := make([]domain.Investor, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, nil
}

func (r *repository) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	var model models.Investor
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("investor with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get investor", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	if inv == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "investor is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Investor{}).
		Where("id = ?", inv.ID).
		Updates(models.FromDomain(inv))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update investor", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("investor with id %d does not exist", inv.ID), nil)
	}
	return nil
}

func (r *repository) DeleteInvestor(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Investor{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete investor", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("investor with id %d does not exist", id), nil)
	}
	return nil
}
