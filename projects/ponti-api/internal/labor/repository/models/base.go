package models

import (
	catmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

type Labor struct {
	ID              int64           `gorm:"primaryKey;autoIncrement"`
	Name            string          `gorm:"type:varchar(255);not null;column:name"`
	ContractorName  string          `gorm:"type:varchar(255);not null;column:contractor_name"`
	Price           decimal.Decimal `gorm:"not null;column:price"`
	ProjectId       int64           `gorm:"not null;column:project_id"`
	LaborCategoryID int64           `gorm:"not null;column:category_id"`
	Category        catmod.Category `gorm:"foreignKey:LaborCategoryID;references:ID" json:"category"`

	sharedmodels.Base
}

func (l Labor) ToDomain() *domain.Labor {
	return &domain.Labor{
		ID:             l.ID,
		Name:           l.Name,
		ContractorName: l.ContractorName,
		Price:          l.Price,
		ProjectId:      l.ProjectId,
		CategoryId:     l.LaborCategoryID,
	}
}

func FromDomain(d *domain.Labor) *Labor {
	return &Labor{
		ID:              d.ID,
		Name:            d.Name,
		ContractorName:  d.ContractorName,
		Price:           d.Price,
		ProjectId:       d.ProjectId,
		LaborCategoryID: d.CategoryId,
		Base: sharedmodels.Base{
			CreatedBy: d.Base.CreatedBy,
			UpdatedBy: d.Base.UpdatedBy,
		},
	}
}
