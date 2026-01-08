package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	decimal "github.com/shopspring/decimal"
)

type CropCommercialization struct {
	ID             int64           `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID      int64           `gorm:"not null;index;column:project_id"`
	CropID         int64           `gorm:"not null;column:crop_id"`
	BoardPrice     decimal.Decimal `gorm:"type:numeric(12,2);not null;column:board_price"`
	FreightCost    decimal.Decimal `gorm:"type:numeric(12,2);not null;column:freight_cost"`
	CommercialCost decimal.Decimal `gorm:"type:numeric(12,2);not null;column:commercial_cost"`
	NetPrice       decimal.Decimal `gorm:"type:numeric(12,2);not null;column:net_price"`

	sharedmodels.Base
}

func FromDomain(cc *domain.CropCommercialization) *CropCommercialization {
	return &CropCommercialization{
		ID:             cc.ID,
		ProjectID:      cc.ProjectID,
		CropID:         cc.CropID,
		BoardPrice:     cc.BoardPrice,
		FreightCost:    cc.FreightCost,
		CommercialCost: cc.CommercialCost,
		NetPrice:       cc.NetPrice,
		Base: sharedmodels.Base{
			CreatedBy: cc.CreatedBy,
			UpdatedBy: cc.UpdatedBy,
		},
	}
}

func (m *CropCommercialization) ToDomain() *domain.CropCommercialization {
	return &domain.CropCommercialization{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		CropID:         m.CropID,
		BoardPrice:     m.BoardPrice,
		FreightCost:    m.FreightCost,
		CommercialCost: m.CommercialCost,
		NetPrice:       m.NetPrice,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedAt: m.UpdatedAt,
			UpdatedBy: m.UpdatedBy,
		},
	}
}
