package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Labor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          decimal.Decimal
	ProjectId      int64
	CategoryId     int64
	shareddomain.Base
}

type ListedLabor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          decimal.Decimal
	ProjectId      int64
	CategoryId     int64
	CategoryName   string

	shareddomain.Base
}
