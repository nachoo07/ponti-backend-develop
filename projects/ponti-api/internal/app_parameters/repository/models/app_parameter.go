package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/app_parameters/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type AppParameter struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Key         string `gorm:"uniqueIndex;size:100;not null"`
	Value       string `gorm:"size:255;not null"`
	Type        string `gorm:"size:20;not null"` // decimal, integer, string, boolean
	Category    string `gorm:"size:50;not null"` // units, calculations, business_rules
	Description string `gorm:"type:text"`

	sharedmodels.Base
}

func (m *AppParameter) ToDomain() *domain.AppParameter {
	return &domain.AppParameter{
		ID:          m.ID,
		Key:         m.Key,
		Value:       m.Value,
		Type:        m.Type,
		Category:    m.Category,
		Description: m.Description,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.AppParameter) *AppParameter {
	return &AppParameter{
		ID:          d.ID,
		Key:         d.Key,
		Value:       d.Value,
		Type:        d.Type,
		Category:    d.Category,
		Description: d.Description,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
