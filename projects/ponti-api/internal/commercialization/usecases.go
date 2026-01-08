package commercialization

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
)

type RepositoryPort interface {
	CreateBulk(context.Context, []domain.CropCommercialization) error
	ListByProject(context.Context, int64) ([]domain.CropCommercialization, error)
	Update(context.Context, *domain.CropCommercialization) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateOrUpdateBulk(ctx context.Context, items []domain.CropCommercialization) error {
	if len(items) == 0 {
		return types.NewError(types.ErrInvalidInput, "no items provided", nil)
	}

	for i := range items {
		items[i].NetPrice = items[i].CalculateNetPrice()
	}

	var toCreate []domain.CropCommercialization
	for _, it := range items {
		if it.ID == 0 {
			toCreate = append(toCreate, it)
		} else {
			if err := u.repo.Update(ctx, &it); err != nil {
				return types.NewError(types.ErrInternal, "failed to update crop commercialization", err)
			}
		}
	}

	if len(toCreate) > 0 {
		if err := u.repo.CreateBulk(ctx, toCreate); err != nil {
			return types.NewError(types.ErrInternal, "failed to create crop commercialization", err)
		}
	}

	return nil
}

func (u *UseCases) ListByProject(ctx context.Context, projectId int64) ([]domain.CropCommercialization, error) {
	return u.repo.ListByProject(ctx, projectId)
}
