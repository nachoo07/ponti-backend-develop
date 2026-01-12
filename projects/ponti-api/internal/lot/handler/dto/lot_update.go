package dto

import (
	"time"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

// LotUpdate es un DTO específico para actualizaciones de lotes
// No tiene validaciones binding requeridas para permitir actualizaciones parciales
type LotUpdate struct {
	Name             string          `json:"name,omitempty"`
	FieldID          int64           `json:"field_id,omitempty"`
	Hectares         decimal.Decimal `json:"hectares,omitempty"`
	PreviousCropID   int64           `json:"previous_crop_id,omitempty"`
	PreviousCropName string          `json:"previous_crop_name,omitempty"`
	CurrentCropID    int64           `json:"current_crop_id,omitempty"`
	CurrentCropName  string          `json:"current_crop_name,omitempty"`
	Season           string          `json:"season,omitempty"`
	Variety          string          `json:"variety,omitempty"`
	Dates            []LotDates      `json:"dates,omitempty"`
	Status           string          `json:"status,omitempty"`

	// Campos de auditoría
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	UpdatedBy *int64    `json:"updated_by,omitempty"`
}

// ToDomain convierte el DTO de actualización al dominio
func (d *LotUpdate) ToDomain() (*domain.Lot, error) {
	dates := make([]domain.LotDates, len(d.Dates))

	for i, date := range d.Dates {
		var sowingDatePtr *time.Time
		if date.SowingDate != "" {
			sowingDate, err := time.Parse("2006-01-02", date.SowingDate)
			if err != nil {
				return nil, err
			}
			sowingDatePtr = &sowingDate
		}

		var harvestDatePtr *time.Time
		if date.HarvestDate != "" {
			harvestDate, err := time.Parse("2006-01-02", date.HarvestDate)
			if err != nil {
				return nil, err
			}
			harvestDatePtr = &harvestDate
		}

		dates[i] = domain.LotDates{
			SowingDate:  sowingDatePtr,
			HarvestDate: harvestDatePtr,
			Sequence:    date.Sequence,
		}
	}

	// Solo crear objetos Crop si se envían explícitamente
	var previousCrop, currentCrop cropdom.Crop
	if d.PreviousCropID > 0 {
		previousCrop = cropdom.Crop{ID: d.PreviousCropID, Name: d.PreviousCropName}
	}
	if d.CurrentCropID > 0 {
		currentCrop = cropdom.Crop{ID: d.CurrentCropID, Name: d.CurrentCropName}
	}

	lot := &domain.Lot{
		Name:         d.Name,
		FieldID:      d.FieldID,
		Hectares:     d.Hectares,
		PreviousCrop: previousCrop,
		CurrentCrop:  currentCrop,
		Season:       d.Season,
		Variety:      d.Variety,
		Dates:        dates,
		Status:       d.Status,
		Base: shareddomain.Base{
			UpdatedAt: d.UpdatedAt,
		},
	}

	// Preservar UpdatedBy si se proporciona en el DTO
	if d.UpdatedBy != nil {
		lot.Base.UpdatedBy = d.UpdatedBy
	}

	return lot, nil
}
