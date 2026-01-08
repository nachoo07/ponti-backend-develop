package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Category struct {
	ID     int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name   string `gorm:"type:varchar(50);unique;not null"`
	TypeID int64  `gorm:"not null;column:type_id"`

	sharedmodels.Base
}

func (m *Category) ToDomain() *domain.Category {
	return &domain.Category{
		ID:     m.ID,
		Name:   m.Name,
		TypeID: m.TypeID,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Category) *Category {
	return &Category{
		ID:     d.ID,
		Name:   d.Name,
		TypeID: d.TypeID,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
