package domain

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	projdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	supplydomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	supplymovementdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type Stock struct {
	ID               int64
	Project          *projdom.Project
	Supply           *supplydomain.Supply
	Investor         *domain.Investor
	CloseDate        *time.Time
	SupplyMovements  []supplymovementdomain.SupplyMovement
	RealStockUnits   decimal.Decimal
	Consumed         decimal.Decimal
	UnitsTransferred decimal.Decimal
	InitialStock     decimal.Decimal
	YearPeriod       int64
	MonthPeriod      int64
	shareddomain.Base
}

func (s *Stock) GetTotalUSD() decimal.Decimal {
	return s.GetStockUnits().Mul(s.Supply.Price)
}

func (s *Stock) GetStockUnits() decimal.Decimal {

	return s.GetEntryStock().Sub(s.Consumed)
}

func (s *Stock) GetStockDifference() decimal.Decimal {
	return s.RealStockUnits.Sub(s.GetStockUnits())
}

func (s *Stock) GetEntryStock() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if supplyMovement.IsEntry {
			stockUnits = stockUnits.Add(supplyMovement.Quantity)
		}
	}

	if s.UnitsTransferred.GreaterThan(decimal.Zero) {
		stockUnits = stockUnits.Sub(s.UnitsTransferred)
	}

	return stockUnits
}

func (s *Stock) GetOutStock() decimal.Decimal {
	var stockUnits decimal.Decimal
	for _, supplyMovement := range s.SupplyMovements {
		if !supplyMovement.IsEntry {
			stockUnits = stockUnits.Add(supplyMovement.Quantity)
		}
	}

	return stockUnits
}

func (s *Stock) GetSupplyUnitName() string {
	return s.Supply.UnitName
}
