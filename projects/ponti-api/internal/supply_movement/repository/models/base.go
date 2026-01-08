package models

import (
	"time"

	investormod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	provmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/repository/models"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	supplymod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type SupplyMovement struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement;column:id"`
	StockId              int64     `gorm:"not null;column:stock_id"`
	Quantity             decimal.Decimal   `gorm:"not null;column:quantity"`
	MovementType         string    `gorm:"type:text;check:movement_type IN ('Stock','Movimiento interno','Remito oficial', 'Movimiento interno entrada');not null;column:movement_type"`
	MovementDate         *time.Time `gorm:"not null;column:movement_date"`
	ReferenceNumber      string    `gorm:"not null;column:reference_number"`
	ProjectId            int64     `gorm:"not null;column:project_id"`
	ProjectDestinationId int64     `gorm:"not null;column:project_destination_id"`
	SupplyID             int64     `gorm:"not null;column:supply_id"`
	InvestorID           int64     `gorm:"not null;column:investor_id"`
	ProviderID           int64     `gorm:"not null;column:provider_id"`
	IsEntry			 bool      `gorm:"not null;column:is_entry"`

	Supply   supplymod.Supply     `gorm:"foreignKey:SupplyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Investor investormod.Investor `gorm:"foreignKey:InvestorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Provider provmod.Provider     `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	sharedmodels.Base
}

func (s *SupplyMovement) ToDomain() *domain.SupplyMovement {
	return &domain.SupplyMovement{
		ID:                   s.ID,
		StockId:              s.StockId,
		Quantity:             s.Quantity,
		MovementType:         s.MovementType,
		MovementDate:         s.MovementDate,
		ReferenceNumber:      s.ReferenceNumber,
		ProjectId:            s.ProjectId,
		ProjectDestinationId: s.ProjectDestinationId,
		Supply:               s.Supply.ToDomain(),
		Investor:             s.Investor.ToDomain(),
		Provider:             s.Provider.ToDomain(),
		IsEntry: 			  s.IsEntry,
		Base: shareddomain.Base{
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
			CreatedBy: s.CreatedBy,
			UpdatedBy: s.UpdatedBy,
		},
	}

}

func FromDomain(s *domain.SupplyMovement) *SupplyMovement {
	return &SupplyMovement{
		ID:                   s.ID,
		StockId:              s.StockId,
		Quantity:             s.Quantity,
		MovementType:         s.MovementType,
		MovementDate:         s.MovementDate,
		ReferenceNumber:            s.ReferenceNumber,
		ProjectId:            s.ProjectId,
		ProjectDestinationId: s.ProjectDestinationId,
		SupplyID:             s.Supply.ID,
		InvestorID:           s.Investor.ID,
		ProviderID:           s.Provider.ID,
		IsEntry: 			  s.IsEntry,
		Base: sharedmodels.Base{
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
			CreatedBy: s.CreatedBy,
			UpdatedBy: s.UpdatedBy,
		},
	}
}
