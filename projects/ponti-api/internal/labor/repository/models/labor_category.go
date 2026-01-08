package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type LaborCategory struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"not null;column:name"`
	LaborTypeId int64     `gorm:"not null;column:type_id"`
	LaborType   LaborType `gorm:"foreignKey:type_id;references:ID"`

	sharedmodels.Base
}

type LaborType struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;column:name"`

	sharedmodels.Base
}

func (lc LaborCategory) ToDomain() *domain.LaborCategory {
	return &domain.LaborCategory{
		ID:   lc.ID,
		Name: lc.Name,
		LaborType: domain.LaborType{
			ID:   lc.LaborTypeId,
			Name: lc.LaborType.Name,
		},
	}
}

func LaborCategoryFromDomain(d *domain.LaborCategory) *LaborCategory {
	return &LaborCategory{
		ID:          d.ID,
		Name:        d.Name,
		LaborTypeId: d.LaborTypeId,
		LaborType: LaborType{
			ID:   d.LaborType.ID,
			Name: d.LaborType.Name,
		},
	}
}
