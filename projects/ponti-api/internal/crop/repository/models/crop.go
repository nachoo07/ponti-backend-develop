package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Crop struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"uniqueIndex:idx_crops_name;size:50;not null"`

	sharedmodels.Base
}

func (Crop) TableName() string {
	return "crops"
}

func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:   c.ID,
		Name: c.Name,
		Base: shareddomain.Base{
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		},
	}
}

func FromDomainCrop(d *domain.Crop) *Crop {
	return &Crop{
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
