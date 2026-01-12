package dto

import (
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// Lot matches the POST/PUT payload and includes FieldID.
type Lot struct {
	ID             int64   `json:"id,omitempty"`
	Name           string  `json:"name"`
	FiendID        int64   `json:"field_id"`
	Hectares       float64 `json:"hectares"`
	PreviousCropID int64   `json:"previous_crop_id"`
	CurrentCropID  int64   `json:"current_crop_id"`
	Season         string  `json:"season"`
}

// ToDomain converts the DTO into a domain.Lot.
func (p Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:           p.ID,
		Name:         p.Name,
		FieldID:      p.FiendID,
		Hectares:     p.Hectares,
		PreviousCrop: cropdom.Crop{ID: p.PreviousCropID},
		CurrentCrop:  cropdom.Crop{ID: p.CurrentCropID},
		Season:       p.Season,
	}
}

// FromDomain converts a domain.Lot into a DTO.
func FromDomain(d domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		FiendID:        d.FieldID,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCrop.ID,
		CurrentCropID:  d.CurrentCrop.ID,
		Season:         d.Season,
	}
}
