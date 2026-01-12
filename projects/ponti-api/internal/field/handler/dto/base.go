package dto

import (
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// Field represents a field payload with its related lots.
type Field struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name" binding:"required"`
	LeaseTypeID int64  `json:"lease_type_id" binding:"required"`
	Lots        []Lot  `json:"lots" binding:"required,dive,required"`
}

// Lot represents a lot within a field payload.
type Lot struct {
	Name           string  `json:"name" binding:"required"`
	Hectares       float64 `json:"hectares" binding:"required"`
	PreviousCropID int64   `json:"previous_crop_id" binding:"required"`
	CurrentCropID  int64   `json:"current_crop_id" binding:"required"`
	Season         string  `json:"season" binding:"required"`
}

// ToDomain converts the Field DTO to a domain.Field, including nested lots.
func (f Field) ToDomain() *fielddom.Field {
	d := &fielddom.Field{
		ID:          f.ID,
		Name:        f.Name,
		LeaseTypeID: f.LeaseTypeID,
	}
	for _, lt := range f.Lots {
		d.Lots = append(d.Lots, lotdom.Lot{
			Name:         lt.Name,
			Hectares:     lt.Hectares,
			PreviousCrop: cropdom.Crop{ID: lt.PreviousCropID},
			CurrentCrop:  cropdom.Crop{ID: lt.CurrentCropID},
			Season:       lt.Season,
		})
	}
	return d
}

// FromDomain converts a domain.Field to the Field DTO, including nested lots.
func FromDomain(d fielddom.Field) Field {
	r := Field{
		ID:          d.ID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseTypeID,
	}
	for _, ld := range d.Lots {
		r.Lots = append(r.Lots, Lot{
			Name:           ld.Name,
			Hectares:       ld.Hectares,
			PreviousCropID: ld.PreviousCrop.ID,
			CurrentCropID:  ld.CurrentCrop.ID,
			Season:         ld.Season,
		})
	}
	return r
}
