package models

import (
	"time"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type Field struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID   int64     `gorm:"index;column:project_id"`
	Name        string    `gorm:"size:100;not null;column:name"`
	LeaseTypeID int64     `gorm:"not null;index;column:lease_type_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at"`
	Lots        []Lot     `gorm:"foreignKey:FieldID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Lot struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id"`
	FieldID        int64     `gorm:"not null;index;column:field_id"`
	Name           string    `gorm:"size:100;not null;column:name"`
	Hectares       float64   `gorm:"not null;column:hectares"`
	PreviousCropID int64     `gorm:"not null;column:previous_crop_id"`
	CurrentCropID  int64     `gorm:"not null;column:current_crop_id"`
	Season         string    `gorm:"size:20;not null;column:season"`
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (m Field) ToDomain() *domain.Field {
	d := &domain.Field{
		ID:          m.ID,
		Name:        m.Name,
		LeaseTypeID: m.LeaseTypeID,
	}
	for _, lotModel := range m.Lots {
		d.Lots = append(d.Lots, lotModel.ToDomain())
	}
	return d
}

func (m Lot) ToDomain() lotdom.Lot {
	return lotdom.Lot{
		ID:           m.ID,
		Name:         m.Name,
		Hectares:     m.Hectares,
		PreviousCrop: cropdom.Crop{ID: m.PreviousCropID},
		CurrentCrop:  cropdom.Crop{ID: m.CurrentCropID},
		Season:       m.Season,
	}
}

func FromDomain(d *domain.Field) *Field {
	m := &Field{
		ID:          d.ID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseTypeID,
	}
	for _, ld := range d.Lots {
		m.Lots = append(m.Lots, Lot{
			FieldID:        d.ID,
			Name:           ld.Name,
			Hectares:       ld.Hectares,
			PreviousCropID: ld.PreviousCrop.ID,
			CurrentCropID:  ld.CurrentCrop.ID,
			Season:         ld.Season,
		})
	}
	return m
}
