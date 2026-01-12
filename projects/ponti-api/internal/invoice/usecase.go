package invoice

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/usecases/domain"
)

type RepositoryPort interface {
	GetByWorkOrderID(context.Context, int64) (*domain.Invoice, error)
	Create(context.Context, *domain.Invoice) (int64, error)
	Update(context.Context, *domain.Invoice) error
	Delete(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetInvoiceByWorkOrder(ctx context.Context, WorkOrderID int64) (*domain.Invoice, error) {
	return u.repo.GetByWorkOrderID(ctx, WorkOrderID)
}

func (u *UseCases) CreateInvoice(ctx context.Context, item *domain.Invoice) (int64, error) {
	return u.repo.Create(ctx, item)
}

func (u *UseCases) UpdateInvoice(ctx context.Context, item *domain.Invoice) error {
	return u.repo.Update(ctx, item)
}

func (u *UseCases) DeleteInvoice(ctx context.Context, WorkOrderID int64) error {
	return u.repo.Delete(ctx, WorkOrderID)
}
