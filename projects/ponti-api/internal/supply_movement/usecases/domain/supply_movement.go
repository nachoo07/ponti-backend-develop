package domain

import (
	"time"

	investordomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	provaderdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	suplydomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

type SupplyMovement struct {
	ID                   int64
	StockId              int64
	Quantity             decimal.Decimal
	MovementType         string
	MovementDate         *time.Time
	ReferenceNumber      string
	ProjectId            int64
	ProjectDestinationId int64
	Supply               *suplydomain.Supply
	Investor             *investordomain.Investor
	Provider             *provaderdomain.Provider
	IsEntry 			 bool
	shareddomain.Base
}
