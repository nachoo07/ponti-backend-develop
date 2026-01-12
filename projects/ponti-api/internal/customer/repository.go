package customer

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository crea una instancia del adaptador GORM para customers.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	if c == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "customer is nil", nil)
	}
	model := models.FromDomain(c)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create customer", err)
	}
	return model.ID, nil
}

func (r *repository) ListCustomers(ctx context.Context) ([]domain.Customer, error) {
	var list []models.Customer
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list customers", err)
	}
	result := make([]domain.Customer, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *repository) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	var model models.Customer
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("customer with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get customer", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	if c == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "customer is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", c.ID).
		Updates(models.FromDomain(c))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update customer", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *repository) DeleteCustomer(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Customer{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete customer", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", id), nil)
	}
	return nil
}
