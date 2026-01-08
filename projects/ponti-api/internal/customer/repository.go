package customer

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
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

func (r *Repository) CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	if c == nil {
		return 0, types.NewError(types.ErrValidation, "customer is nil", nil)
	}
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create customer", err)
	}
	return model.ID, nil
}

func (r *Repository) ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var list []models.Customer
	var total int64

	db0 := r.db.Client().WithContext(ctx).Model(&models.Customer{})

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count customers", err)
	}

	// Consulta ligera: sólo id y name
	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list customers", err)
	}

	// Mapear a dominio ligero
	customers := make([]domain.ListedCustomer, len(list))
	for i, m := range list {
		customers[i] = domain.ListedCustomer{
			ID:   m.ID,
			Name: m.Name,
		}
	}

	return customers, total, nil
}

func (r *Repository) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	var model models.Customer
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get customer", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	if c == nil {
		return types.NewError(types.ErrValidation, "customer is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", c.ID).
		Updates(models.FromDomain(c))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update customer", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteCustomer(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Customer{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete customer", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d does not exist", id), nil)
	}
	return nil
}
