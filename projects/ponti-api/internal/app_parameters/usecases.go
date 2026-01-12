package app_parameters

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/app_parameters/usecases/domain"
)

type RepositoryPort interface {
	GetByKey(ctx context.Context, key string) (*domain.AppParameter, error)
	ListByCategory(ctx context.Context, category string) ([]domain.AppParameter, error)
	ListAll(ctx context.Context) ([]domain.AppParameter, error)
	Create(ctx context.Context, item *domain.AppParameter) (int64, error)
	Update(ctx context.Context, item *domain.AppParameter) error
	Delete(ctx context.Context, id int64) error
}

type UseCases struct {
	repository RepositoryPort
}

func NewUseCases(repository RepositoryPort) *UseCases {
	return &UseCases{
		repository: repository,
	}
}

func (u *UseCases) GetParameter(ctx context.Context, key string) (*domain.AppParameter, error) {
	if key == "" {
		return nil, types.NewError(types.ErrInvalidInput, "key is required", nil)
	}

	return u.repository.GetByKey(ctx, key)
}

func (u *UseCases) GetParametersByCategory(ctx context.Context, category string) ([]domain.AppParameter, error) {
	if category == "" {
		return nil, types.NewError(types.ErrInvalidInput, "category is required", nil)
	}

	return u.repository.ListByCategory(ctx, category)
}

func (u *UseCases) GetAllParameters(ctx context.Context) ([]domain.AppParameter, error) {
	return u.repository.ListAll(ctx)
}

func (u *UseCases) CreateParameter(ctx context.Context, param *domain.AppParameter) (int64, error) {
	if param.Key == "" || param.Value == "" || param.Type == "" || param.Category == "" {
		return 0, types.NewError(types.ErrInvalidInput, "missing required fields", nil)
	}

	return u.repository.Create(ctx, param)
}

func (u *UseCases) UpdateParameter(ctx context.Context, param *domain.AppParameter) error {
	if param.ID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid id", nil)
	}

	if param.Key == "" || param.Value == "" || param.Type == "" || param.Category == "" {
		return types.NewError(types.ErrInvalidInput, "missing required fields", nil)
	}

	return u.repository.Update(ctx, param)
}

func (u *UseCases) DeleteParameter(ctx context.Context, id int64) error {
	if id == 0 {
		return types.NewError(types.ErrInvalidID, "invalid id", nil)
	}

	return u.repository.Delete(ctx, id)
}
