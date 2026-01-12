package models

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

// Crop represents a type of crop.
type Crop struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

// ToDomain converts the Crop model to the domain entity.
func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:   c.ID,
		Name: c.Name,
	}
}

// FromDomainCrop converts a domain Crop entity to the GORM model.
func FromDomainCrop(d *domain.Crop) *Crop {
	return &Crop{
		ID:   d.ID,
		Name: d.Name,
	}
}
