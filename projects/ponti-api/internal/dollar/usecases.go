package dollar

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type RepositoryPort interface {
	ListByProject(context.Context, int64) ([]domain.DollarAverage, error)
	Create(context.Context, *domain.DollarAverage) (int64, error)
	Update(context.Context, *domain.DollarAverage) error
	GetByComposite(ctx context.Context, projectID, year int64, month string) (*domain.DollarAverage, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) ListByProject(ctx context.Context, projectID int64) ([]domain.DollarAverage, error) {
	if projectID == 0 {
		return nil, types.NewError(types.ErrInvalidID, "projectID is required", nil)
	}
	return u.repo.ListByProject(ctx, projectID)
}

func (u *UseCases) CreateOrUpdateBulk(ctx context.Context, items []domain.DollarAverage) error {
	if len(items) == 0 {
		return types.NewError(types.ErrInternal, "no values provided", nil)
	}

	base := items[0]
	seen := make(map[string]bool)

	for _, d := range items {
		if d.ProjectID != base.ProjectID || d.Year != base.Year {
			return types.NewError(types.ErrBadRequest, "all items must share projectID and year", nil)
		}
		if seen[d.Month] {
			return types.NewError(types.ErrValidation, "duplicate month", nil)
		}
		seen[d.Month] = true

		existing, err := u.repo.GetByComposite(ctx, d.ProjectID, d.Year, d.Month)
		if err != nil {
			return types.NewError(types.ErrInternal, "error checking existence for month", err)
		}

		if existing == nil {
			id, err := u.repo.Create(ctx, &d)
			if err != nil {
				return types.NewError(types.ErrInternal, "error creating month", err)
			}
			d.ID = id
		} else {
			d.ID = existing.ID
			if err := u.repo.Update(ctx, &d); err != nil {
				return types.NewError(types.ErrInternal, "error updating month", err)
			}
		}
	}

	return nil
}
