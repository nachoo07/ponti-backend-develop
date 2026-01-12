package models

import (
	"time"

	cropmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

type Lot struct {
	ID             int64           `gorm:"primaryKey;autoIncrement;column:id"`
	Name           string          `gorm:"type:varchar(100);not null;column:name"`
	FieldID        int64           `gorm:"not null;index;constraint:OnDelete:CASCADE;column:field_id"`
	Hectares       decimal.Decimal `gorm:"type:numeric(18,6);not null;column:hectares"`
	PreviousCropID int64           `gorm:"not null;index;column:previous_crop_id"`
	CurrentCropID  int64           `gorm:"not null;index;column:current_crop_id"`
	Season         string          `gorm:"size:20;not null;column:season"`
	Variety        string          `gorm:"type:text;column:variety"`
	SowingDate     *time.Time      `gorm:"type:date;column:sowing_date"`
	Tons           decimal.Decimal `gorm:"type:numeric;column:tons"`

	sharedmodels.Base // <-- embebe campos GORM de auditoría

	PreviousCrop cropmod.Crop `gorm:"foreignKey:PreviousCropID;references:ID"`
	CurrentCrop  cropmod.Crop `gorm:"foreignKey:CurrentCropID;references:ID"`
	Dates        []LotDates   `gorm:"foreignKey:LotID;references:ID"`
}

// ToDomain convierte el modelo Lot en su equivalente de dominio.
func (m *Lot) ToDomain() *domain.Lot {
	// Convertir las fechas del modelo a dominio
	domainDates := make([]domain.LotDates, 0, len(m.Dates))
	for _, date := range m.Dates {
		domainDates = append(domainDates, domain.LotDates{
			SowingDate:  date.SowingDate,
			HarvestDate: date.HarvestDate,
			Sequence:    date.Sequence,
		})
	}

	return &domain.Lot{
		ID:           m.ID,
		Name:         m.Name,
		FieldID:      m.FieldID,
		Hectares:     m.Hectares,
		PreviousCrop: cropdom.Crop{ID: m.PreviousCropID},
		CurrentCrop:  cropdom.Crop{ID: m.CurrentCropID},
		Season:       m.Season,
		Variety:      m.Variety,
		SowingDate:   m.SowingDate,
		Tons:         m.Tons,
		Dates:        domainDates,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

// FromDomain convierte la entidad de dominio Lot en su modelo de persistencia.
func FromDomain(d *domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		FieldID:        d.FieldID,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCrop.ID,
		CurrentCropID:  d.CurrentCrop.ID,
		Season:         d.Season,
		Variety:        d.Variety,
		SowingDate:     d.SowingDate,
		Tons:           d.Tons,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
