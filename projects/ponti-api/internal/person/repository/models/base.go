package models

import (
	"errors"
	"time"

	"github.com/lib/pq" // Para manejar arrays de texto en PostgreSQL.
	"gorm.io/gorm"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/usecases/domain"
)

type Person struct {
	ID         string         `gorm:"primaryKey;column:id"`
	FirstName  string         `gorm:"column:first_name"`
	LastName   string         `gorm:"column:last_name"`
	Age        int            `gorm:"column:age"`
	Gender     string         `gorm:"column:gender"`
	NationalID int64          `gorm:"column:national_id;uniqueIndex"` // Restricción única.
	Phone      string         `gorm:"column:phone"`
	Interests  pq.StringArray `gorm:"type:text[];column:interests"` // Array de texto.
	Hobbies    pq.StringArray `gorm:"type:text[];column:hobbies"`   // Array de texto.
	Deleted    bool           `gorm:"column:deleted"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Person) TableName() string {
	return "people"
}

func FromDomain(person *domain.Person) (*Person, error) {
	if person == nil {
		return nil, errors.New("person cannot be nil")
	}

	return &Person{
		ID:         person.ID,
		FirstName:  person.FirstName,
		LastName:   person.LastName,
		Age:        person.Age,
		Gender:     person.Gender,
		NationalID: person.NationalID,
		Phone:      person.Phone,
		Interests:  pq.StringArray(person.Interests),
		Hobbies:    pq.StringArray(person.Hobbies),
		Deleted:    person.Deleted,
	}, nil
}

func (pm *Person) ToDomain() (*domain.Person, error) {
	if pm == nil {
		return nil, errors.New("models.Person is nil")
	}

	return &domain.Person{
		ID:         pm.ID,
		FirstName:  pm.FirstName,
		LastName:   pm.LastName,
		Age:        pm.Age,
		Gender:     pm.Gender,
		NationalID: pm.NationalID,
		Phone:      pm.Phone,
		Interests:  []string(pm.Interests),
		Hobbies:    []string(pm.Hobbies),
		Deleted:    pm.Deleted,
	}, nil
}
