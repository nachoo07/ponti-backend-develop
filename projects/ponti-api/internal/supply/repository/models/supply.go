package models

import (
	catmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	classtype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	classdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

// Modelo principal de Supply
type Supply struct {
	ID        int64           `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID int64           `gorm:"not null;index"`
	Name      string          `gorm:"type:varchar(100);not null"`
	Price     decimal.Decimal `gorm:"not null"`

	UnitID int64

	CategoryID int64
	Category   catmod.Category `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	TypeID int64
	Type   classtype.ClassType `gorm:"foreignKey:TypeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	sharedmodels.Base // Campos de auditoría (CreatedAt, UpdatedAt, etc)
}

// De persistencia (models.Supply) → dominio (domain.Supply)
func (m *Supply) ToDomain() *domain.Supply {
	// Obtener nombre de unidad desde app_parameters
	unitName := m.getUnitName()

	return &domain.Supply{
		ID:           m.ID,
		ProjectID:    m.ProjectID,
		Name:         m.Name,
		UnitID:       int64(m.UnitID),
		Price:        m.Price,
		CategoryID:   int64(m.CategoryID),
		CategoryName: m.Category.Name,
		Type: classdomain.ClassType{
			ID:   int64(m.TypeID),
			Name: m.Type.Name,
		},
		UnitName: unitName,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Supply) *Supply {
	return &Supply{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		Name:       d.Name,
		Price:      d.Price,
		UnitID:     int64(d.UnitID),
		CategoryID: int64(d.CategoryID),
		TypeID:     int64(d.Type.ID),
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

// getUnitName obtiene el nombre de la unidad desde app_parameters
func (m *Supply) getUnitName() string {
	// Mapeo temporal hasta que se implemente la consulta a app_parameters
	// TODO: Implementar consulta real a app_parameters
	switch m.UnitID {
	case 1:
		return "Lt" // unit_liters
	case 2:
		return "Kg" // unit_kilos
	case 3:
		return "Ha" // unit_hectares
	default:
		return "Unknown"
	}
}
