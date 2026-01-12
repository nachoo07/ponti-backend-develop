package customer

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases crea una instancia de los casos de uso para Customer.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	return u.repo.CreateCustomer(ctx, c)
}

func (u *useCases) ListCustomers(ctx context.Context) ([]domain.Customer, error) {
	return u.repo.ListCustomers(ctx)
}

func (u *useCases) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	return u.repo.GetCustomer(ctx, id)
}

func (u *useCases) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	return u.repo.UpdateCustomer(ctx, c)
}

func (u *useCases) DeleteCustomer(ctx context.Context, id int64) error {
	return u.repo.DeleteCustomer(ctx, id)
}
