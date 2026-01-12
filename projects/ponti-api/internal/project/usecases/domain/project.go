package domain

import (
	campdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fieldom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Project struct {
	ID                 int64
	Name               string
	AdminCost          decimal.Decimal
	PlannedCost        decimal.Decimal
	Customer           customerdom.Customer
	Campaign           campdom.Campaign
	Managers           []managerdom.Manager
	Investors          []investordom.Investor
	AdminCostInvestors []investordom.Investor
	Fields             []fieldom.Field

	shareddomain.Base
}

type ListedProject struct {
	ID   int64
	Name string
}
