package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type DollarAverage struct {
	ID         int64
	ProjectID  int64
	Year       int64
	Month      string
	StartValue decimal.Decimal
	EndValue   decimal.Decimal
	AvgValue   decimal.Decimal

	shareddomain.Base // CreatedAt, UpdatedAt, CreatedBy, UpdatedBy, DeletedBy
}
