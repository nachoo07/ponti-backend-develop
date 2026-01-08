package models

import (
	"time"

	investormod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	projmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	supplymod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	supplymovmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/repository/models"
	supplymovementdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type Stock struct {
	ID              int64                         `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID       int64                         `gorm:"not null;index;column:project_id"`
	Project         projmod.Project               `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	SupplyID        int64                         `gorm:"not null;index;column:supply_id"`
	Supply          supplymod.Supply              `gorm:"foreignKey:SupplyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	InvestorID      int64                         `gorm:"not null;index;column:investor_id"`
	Investor        investormod.Investor          `gorm:"foreignKey:InvestorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CloseDate       *time.Time                    `gorm:"null;column:close_date"`
	SupplyMovements []supplymovmod.SupplyMovement `gorm:"foreignKey:StockId;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	InitialStock    decimal.Decimal               `gorm:"not null;column:initial_units"`
	YearPeriod      int64                         `gorm:"not null;column:year_period"`
	MonthPeriod     int64                         `gorm:"not null;column:month_period"`
	UnitsEntered    decimal.Decimal               `gorm:"not null;column:units_entered"`
	Consumed        decimal.Decimal               `gorm:"-"`
	UnitsConsumed   decimal.Decimal               `gorm:"not null;column:units_consumed"`
	RealStockUnits  decimal.Decimal               `gorm:"not null;column:real_stock_units"`
	sharedmodels.Base
}

// ToDomain convierte el modelo Stock a la entidad de dominio
func (m *Stock) ToDomain() *domain.Stock {

	supplyMovementsDomains := make([]supplymovementdomain.SupplyMovement, len(m.SupplyMovements))

	for i, supplyMovement := range m.SupplyMovements {
		supplyMovementsDomains[i] = *supplyMovement.ToDomain()
	}

	return &domain.Stock{
		ID:               m.ID,
		Project:          m.Project.ToDomain(),
		Supply:           m.Supply.ToDomain(),
		CloseDate:        m.CloseDate,
		RealStockUnits:   m.RealStockUnits,
		YearPeriod:       m.YearPeriod,
		MonthPeriod:      m.MonthPeriod,
		Investor:         m.Investor.ToDomain(),
		SupplyMovements:  supplyMovementsDomains,
		Consumed:         m.Consumed,
		UnitsTransferred: m.UnitsConsumed,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

// FromDomain convierte la entidad de dominio Stock al modelo de persistencia
func FromDomain(d *domain.Stock) *Stock {
	return &Stock{
		ID:              d.ID,
		ProjectID:       d.Project.ID,
		SupplyID:        d.Supply.ID,
		InvestorID:      d.Investor.ID,
		RealStockUnits:  d.RealStockUnits,
		YearPeriod:      d.YearPeriod,
		MonthPeriod:     d.MonthPeriod,
		InitialStock:    d.InitialStock,
		CloseDate:       d.CloseDate,
		SupplyMovements: []supplymovmod.SupplyMovement{},
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
