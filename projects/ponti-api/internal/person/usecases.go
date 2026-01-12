package person

import (
	"context"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/usecases/domain"
)

type useCases struct {
	storage Repository
}

func NewUseCases(s Repository) UseCases {
	return &useCases{
		storage: s,
	}
}

func (u *useCases) CreatePerson(ctx context.Context, person *domain.Person) (string, error) {
	personID, err := u.storage.CreatePerson(ctx, person)
	if err != nil {
		return "", fmt.Errorf("failed to create person: %w", err)
	}
	return personID, nil
}

func (u *useCases) ListPersons(ctx context.Context) ([]domain.Person, error) {
	return u.storage.ListPersons(ctx)
}

func (ps *useCases) GetPerson(ctx context.Context, ID string) (*domain.Person, error) {
	return ps.storage.GetPerson(ctx, ID)
}

func (ps *useCases) UpdatePerson(ctx context.Context, ID string, person *domain.Person) error {
	return ps.storage.UpdatePerson(ctx, ID, person)
}

func (ps *useCases) DeletePerson(ctx context.Context, ID string, hardDelete bool) error {
	return ps.storage.DeletePerson(ctx, ID, hardDelete)
}
