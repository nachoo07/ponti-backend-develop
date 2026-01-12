package field

import (
	"context"
	"fmt"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type useCases struct {
	repo Repository
	lot  lot.UseCases
}

func NewUseCases(repo Repository, lot lot.UseCases) UseCases {
	return &useCases{
		repo: repo,
		lot:  lot,
	}
}

func (u *useCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	// 1) Crear el Field y obtener su ID
	fieldID, err := u.repo.CreateField(ctx, f)
	if err != nil {
		return 0, fmt.Errorf("create field %q: %w", f.Name, err)
	}

	// 2) Crear cada Lot apuntando a ese fieldID
	for _, l := range f.Lots {
		l.FieldID = fieldID
		if _, err := u.lot.CreateLot(ctx, &l); err != nil {
			// Si falla, se borra el Field recién creado
			if delErr := u.repo.DeleteField(ctx, fieldID); delErr != nil {
				return 0, fmt.Errorf(
					"rollback delete field %d failed: %v (original error: %w)",
					fieldID, delErr, err,
				)
			}
			return 0, fmt.Errorf("create lot %q for field %q: %w", l.Name, f.Name, err)
		}
	}

	return fieldID, nil
}

func (u *useCases) ListFields(ctx context.Context) ([]domain.Field, error) {
	fields, err := u.repo.ListFields(ctx)
	if err != nil {
		return nil, err
	}
	for i := range fields {
		if err := u.enrichField(ctx, &fields[i]); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

func (u *useCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	f, err := u.repo.GetField(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.enrichField(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (u *useCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}

func (u *useCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}

// helpers
func (u *useCases) enrichField(ctx context.Context, f *domain.Field) error {
	allLots, err := u.lot.ListLots(ctx)
	if err != nil {
		return fmt.Errorf("listar lots: %w", err)
	}

	var related []lotdom.Lot
	for _, l := range allLots {
		if l.FieldID == f.ID {
			related = append(related, l)
		}
	}
	f.Lots = related
	return nil
}
