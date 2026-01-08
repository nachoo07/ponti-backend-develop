package dto

import (
	domainclasstype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
	"time"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	decimal "github.com/shopspring/decimal"
)

// DTO para entrada y salida (puedes separarlo si quieres)
type Supply struct {
	ID        int64           `json:"id,omitempty"`
	ProjectID int64           `json:"project_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`

	UnitID     int64 `json:"unit_id"`
	CategoryID int64 `json:"category_id"`
	TypeID     int64 `json:"type_id"`

	// Audit fields opcionales
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	UpdatedBy *int64    `json:"updated_by,omitempty"`
}

// Para entrada (request): solo mapear IDs
func (d *Supply) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		Name:       d.Name,
		Price:      d.Price,
		UnitID:     d.UnitID,
		CategoryID: d.CategoryID,
		Type: domainclasstype.ClassType{
			ID: d.TypeID,
		},
		Base: shareddomain.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

// Para salida (response): se pueden agregar los nombres si tienes preload
func FromDomain(s *domain.Supply) *Supply {
	return &Supply{
		ID:         s.ID,
		ProjectID:  s.ProjectID,
		Name:       s.Name,
		Price:      s.Price,
		UnitID:     s.UnitID,
		CategoryID: s.CategoryID,
		TypeID:     s.Type.ID,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
		CreatedBy:  s.CreatedBy,
		UpdatedBy:  s.UpdatedBy,
	}
}
