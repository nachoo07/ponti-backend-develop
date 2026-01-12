package dto

import (
	"time"

	"github.com/shopspring/decimal"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Lot struct {
	ID               int64           `json:"id,omitempty"`
	Name             string          `json:"name"`
	FieldID          int64           `json:"field_id"`
	Hectares         decimal.Decimal `json:"hectares"`
	PreviousCropID   int64           `json:"previous_crop_id"`
	PreviousCropName string          `json:"previous_crop_name,omitempty"`
	CurrentCropID    int64           `json:"current_crop_id"`
	CurrentCropName  string          `json:"current_crop_name,omitempty"`
	Season           string          `json:"season"`
	Variety          string          `json:"variety"`
	Dates            []LotDates      `json:"dates"`
	Status           string          `json:"status"`

	// Campos de auditoría
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	UpdatedBy *int64    `json:"updated_by,omitempty"`
}

type LotDates struct {
	SowingDate  string `json:"sowing_date"`
	HarvestDate string `json:"harvest_date"`
	Sequence    int    `json:"sequence"`
}

func (d *Lot) ToDomain() (*domain.Lot, error) {
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
		} else {
			harvestDatePtr = nil
		}

		dates[i] = domain.LotDates{
			SowingDate:  sowingDatePtr,
			HarvestDate: harvestDatePtr,
			Sequence:    date.Sequence,
		}
	}

	return &domain.Lot{
		ID:           d.ID,
		Name:         d.Name,
		FieldID:      d.FieldID,
		Hectares:     d.Hectares,
		PreviousCrop: cropdom.Crop{ID: d.PreviousCropID, Name: d.PreviousCropName},
		CurrentCrop:  cropdom.Crop{ID: d.CurrentCropID, Name: d.CurrentCropName},
		Season:       d.Season,
		Variety:      d.Variety,
		Dates:        dates,
		Status:       d.Status,
		Base: shareddomain.Base{
			UpdatedAt: d.UpdatedAt,
		},
	}, nil
}

func FromDomain(l *domain.Lot) *Lot {
	return &Lot{
		ID:               l.ID,
		Name:             l.Name,
		FieldID:          l.FieldID,
		Hectares:         l.Hectares,
		PreviousCropID:   l.PreviousCrop.ID,
		PreviousCropName: l.PreviousCrop.Name,
		CurrentCropID:    l.CurrentCrop.ID,
		CurrentCropName:  l.CurrentCrop.Name,
		Season:           l.Season,
		Variety:          l.Variety,
		Status:           l.Status,
		CreatedAt:        l.Base.CreatedAt,
		UpdatedAt:        l.Base.UpdatedAt,
		CreatedBy:        l.Base.CreatedBy,
		UpdatedBy:        l.Base.UpdatedBy,
	}
}
