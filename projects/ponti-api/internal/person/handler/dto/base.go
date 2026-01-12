package dto

import (
	"errors"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/usecases/domain"
)

type Person struct {
	FirstName  string   `json:"first_name" binding:"required"`                     // Nombre.
	LastName   string   `json:"last_name" binding:"required"`                      // Apellido.
	Age        int      `json:"age" binding:"required,gte=0"`                      // Edad.
	Gender     string   `json:"gender" binding:"required,oneof=male female other"` // Género.
	NationalID int64    `json:"national_id" binding:"required"`                    // Identificador Nacional (DNI).
	Phone      string   `json:"phone" binding:"required"`                          // Teléfono.
	Interests  []string `json:"interests" binding:"required"`                      // Áreas de interés.
	Hobbies    []string `json:"hobbies" binding:"required"`                        // Hobbies.
}

func (dto *Person) ToDomain() *domain.Person {
	return &domain.Person{
		FirstName:  dto.FirstName,
		LastName:   dto.LastName,
		Age:        dto.Age,
		Gender:     dto.Gender,
		NationalID: dto.NationalID,
		Phone:      dto.Phone,
		Interests:  dto.Interests,
		Hobbies:    dto.Hobbies,
	}
}

func FromDomain(p *domain.Person) (*Person, error) {
	if p == nil {
		return nil, errors.New("person cannot be nil")
	}

	return &Person{
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Age:        p.Age,
		Gender:     p.Gender,
		NationalID: p.NationalID,
		Phone:      p.Phone,
		Interests:  p.Interests,
		Hobbies:    p.Hobbies,
	}, nil
}
