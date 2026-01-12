package person

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx" // Para errores pgx.ErrNoRows.
	"github.com/lib/pq"

	pgdb "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/postgresql/pgxpool"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/usecases/domain"
)

type postgresRepository struct {
	postgresRepository pgdb.Repository
}

func NewPostgresRepository(r pgdb.Repository) Repository {
	return &postgresRepository{
		postgresRepository: r,
	}
}

func (r *postgresRepository) CreatePerson(ctx context.Context, person *domain.Person) (string, error) {
	if person == nil {
		return "", errors.New("person is nil")
	}

	// Convertir de domain.Person a models.Person.
	model, err := models.FromDomain(person)
	if err != nil {
		return "", err
	}
	// Generar un nuevo ID.
	model.ID = uuid.New().String()

	query := `
		INSERT INTO people (
			id,
			national_id,
			first_name,
			last_name,
			age,
			gender,
			phone,
			interests,
			hobbies,
			deleted,
			created_at,
			updated_at,
			deleted_at
		) VALUES (
			$1,  -- id
			$2,  -- national_id
			$3,  -- first_name
			$4,  -- last_name
			$5,  -- age
			$6,  -- gender
			$7,  -- phone
			$8,  -- interests
			$9, -- hobbies
			$10, -- deleted
			NOW(),
			NOW(),
			NULL
		)
		`

	_, err = r.postgresRepository.Pool().Exec(ctx, query,
		model.ID,
		model.NationalID,
		model.FirstName,
		model.LastName,
		model.Age,
		model.Gender,
		model.Phone,
		model.Interests,
		model.Hobbies,
		model.Deleted,
	)
	if err != nil {
		// Verificar si se trata de una violación de restricción única.
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", errors.New("person already exists")
		}
		return "", fmt.Errorf("error creating person: %w", err)
	}

	return model.ID, nil
}

func (r *postgresRepository) ListPersons(ctx context.Context) ([]domain.Person, error) {
	query := `
		SELECT
			id,
			first_name,
			last_name,
			age,
			gender,
			national_id,
			phone,
			interests,
			hobbies,
			deleted,
			created_at,
			updated_at,
			deleted_at
		FROM people
		`

	rows, err := r.postgresRepository.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying people: %w", err)
	}
	defer rows.Close()

	var people []domain.Person
	for rows.Next() {
		var pm models.Person
		if err := rows.Scan(
			&pm.ID,
			&pm.FirstName,
			&pm.LastName,
			&pm.Age,
			&pm.Gender,
			&pm.NationalID,
			&pm.Phone,
			&pm.Interests,
			&pm.Hobbies,
			&pm.Deleted,
			&pm.CreatedAt,
			&pm.UpdatedAt,
			&pm.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning person row: %w", err)
		}

		personDomain, err := pm.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("error converting person to domain: %w", err)
		}
		people = append(people, *personDomain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating person rows: %w", err)
	}

	return people, nil
}

func (r *postgresRepository) GetPerson(ctx context.Context, id string) (*domain.Person, error) {
	query := `
		SELECT
			id,
			first_name,
			last_name,
			age,
			gender,
			national_id,
			phone,
			interests,
			hobbies,
			deleted,
			created_at,
			updated_at,
			deleted_at
		FROM people
		WHERE id = $1
		`

	var pm models.Person
	err := r.postgresRepository.Pool().QueryRow(ctx, query, id).Scan(
		&pm.ID,
		&pm.FirstName,
		&pm.LastName,
		&pm.Age,
		&pm.Gender,
		&pm.NationalID,
		&pm.Phone,
		&pm.Interests,
		&pm.Hobbies,
		&pm.Deleted,
		&pm.CreatedAt,
		&pm.UpdatedAt,
		&pm.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("person not found")
		}
		if pqErr, ok := err.(*pq.Error); ok {
			return nil, fmt.Errorf("database error: %w", pqErr)
		}
		return nil, fmt.Errorf("error getting person by id: %w", err)
	}

	personDomain, err := pm.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("error converting person to domain: %w", err)
	}

	return personDomain, nil
}

func (r *postgresRepository) UpdatePerson(ctx context.Context, ID string, person *domain.Person) error {
	query := `
		UPDATE people
		SET
			national_id = $1,
			first_name = $2,
			last_name = $3,
			age = $4,
			gender = $5,
			phone = $6,
			interests = $7,
			hobbies = $8,
			deleted = $9,
			updated_at = NOW()
		WHERE id = $10
		`

	result, err := r.postgresRepository.Pool().Exec(ctx, query,
		person.NationalID,
		person.FirstName,
		person.LastName,
		person.Age,
		person.Gender,
		person.Phone,
		person.Interests,
		person.Hobbies,
		person.Deleted,
		ID,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			return fmt.Errorf("database error: %w", pqErr)
		}
		return fmt.Errorf("error updating person: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("person not found")
	}

	return nil
}

func (r *postgresRepository) DeletePerson(ctx context.Context, id string, hardDelete bool) error {
	if hardDelete {
		query := `
			DELETE FROM people
			WHERE id = $1
			`
		result, err := r.postgresRepository.Pool().Exec(ctx, query, id)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				return fmt.Errorf("database error: %w", pqErr)
			}
			return fmt.Errorf("error performing hard delete: %w", err)
		}

		if result.RowsAffected() == 0 {
			return errors.New("person not found")
		}

		return nil
	}

	query := `
		UPDATE people
		SET deleted = true, deleted_at = NOW()
		WHERE id = $1
		`
	result, err := r.postgresRepository.Pool().Exec(ctx, query, id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			return fmt.Errorf("database error: %w", pqErr)
		}
		return fmt.Errorf("error performing soft delete: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("person not found")
	}

	return nil
}
