package dto

import (
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	leasetypedom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	"github.com/shopspring/decimal"
)

// Lot DTO para Field, enriquecido
type Lot struct {
	ID               int64           `json:"id,omitempty"`
	Name             string          `json:"name" binding:"required"`
	Hectares         decimal.Decimal `json:"hectares" binding:"required"`
	PreviousCropID   int64           `json:"previous_crop_id" binding:"required"`
	PreviousCropName string          `json:"previous_crop_name,omitempty"`
	CurrentCropID    int64           `json:"current_crop_id" binding:"required"`
	CurrentCropName  string          `json:"current_crop_name,omitempty"`
	Season           string          `json:"season" binding:"required"`
}

// Field DTO
type Field struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name" binding:"required"`
	LeaseTypeID int64  `json:"lease_type_id" binding:"required"`
	Lots        []Lot  `json:"lots" binding:"required,dive,required"`
}

func (f *Field) ToDomain() *fielddom.Field {
	d := &fielddom.Field{
		ID:        f.ID,
		Name:      f.Name,
		LeaseType: &leasetypedom.LeaseType{ID: f.LeaseTypeID},
	}
	for _, lt := range f.Lots {
		d.Lots = append(d.Lots, lotdom.Lot{
			ID:           lt.ID,
			Name:         lt.Name,
			Hectares:     lt.Hectares,
			PreviousCrop: cropdom.Crop{ID: lt.PreviousCropID, Name: lt.PreviousCropName},
			CurrentCrop:  cropdom.Crop{ID: lt.CurrentCropID, Name: lt.CurrentCropName},
			Season:       lt.Season,
		})
	}
	return d
}

func FromDomain(d fielddom.Field) Field {
	r := Field{
		ID:          d.ID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseType.ID,
	}
	for _, ld := range d.Lots {
		r.Lots = append(r.Lots, Lot{
			ID:               ld.ID,
			Name:             ld.Name,
			Hectares:         ld.Hectares,
			PreviousCropID:   ld.PreviousCrop.ID,
			PreviousCropName: ld.PreviousCrop.Name,
			CurrentCropID:    ld.CurrentCrop.ID,
			CurrentCropName:  ld.CurrentCrop.Name,
			Season:           ld.Season,
		})
	}
	return r
}
