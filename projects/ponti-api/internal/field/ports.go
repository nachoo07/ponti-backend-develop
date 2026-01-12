// File: internal/field/ports.go
package field

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
)

// UseCases defines business operations for Field.
type UseCases interface {
	CreateField(ctx context.Context, f *domain.Field) (int64, error)
	ListFields(ctx context.Context) ([]domain.Field, error)
	GetField(ctx context.Context, id int64) (*domain.Field, error)
	UpdateField(ctx context.Context, f *domain.Field) error
	DeleteField(ctx context.Context, id int64) error
}

// Repository defines persistence operations for Field.
type Repository interface {
	CreateField(ctx context.Context, f *domain.Field) (int64, error)
	ListFields(ctx context.Context) ([]domain.Field, error)
	GetField(ctx context.Context, id int64) (*domain.Field, error)
	UpdateField(ctx context.Context, f *domain.Field) error
	DeleteField(ctx context.Context, id int64) error
}
