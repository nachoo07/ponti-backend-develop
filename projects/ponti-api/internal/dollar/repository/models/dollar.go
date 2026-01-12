package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

type ProjectDollarValue struct {
	ID           int64           `gorm:"primaryKey;autoIncrement"`
	ProjectID    int64           `gorm:"not null;index"`
	Year         int64           `gorm:"not null;index"`
	Month        string          `gorm:"size:20;not null;index"`
	StartValue   decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	EndValue     decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	AverageValue decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	sharedmodels.Base
}

func FromDomain(d *domain.DollarAverage) *ProjectDollarValue {
	return &ProjectDollarValue{
		ID:           d.ID,
		ProjectID:    d.ProjectID,
		Year:         d.Year,
		Month:        d.Month,
		StartValue:   d.StartValue,
		EndValue:     d.EndValue,
		AverageValue: d.AvgValue,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

func (m *ProjectDollarValue) ToDomain() *domain.DollarAverage {
	return &domain.DollarAverage{
		ID:         m.ID,
		ProjectID:  m.ProjectID,
		Year:       m.Year,
		Month:      m.Month,
		StartValue: m.StartValue,
		EndValue:   m.EndValue,
		AvgValue:   m.AverageValue,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}
