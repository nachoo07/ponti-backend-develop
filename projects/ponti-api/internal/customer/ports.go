package customer

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

// UseCases define las operaciones de negocio para Customer.
type UseCases interface {
	CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error)
	ListCustomers(ctx context.Context) ([]domain.Customer, error)
	GetCustomer(ctx context.Context, id int64) (*domain.Customer, error)
	UpdateCustomer(ctx context.Context, c *domain.Customer) error
	DeleteCustomer(ctx context.Context, id int64) error
}

// Repository define las operaciones para Customer.
type Repository interface {
	CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error)
	ListCustomers(ctx context.Context) ([]domain.Customer, error)
	GetCustomer(ctx context.Context, id int64) (*domain.Customer, error)
	UpdateCustomer(ctx context.Context, c *domain.Customer) error
	DeleteCustomer(ctx context.Context, id int64) error
}
