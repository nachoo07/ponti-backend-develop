package domain

import (
	"time"

	"github.com/shopspring/decimal"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

// Workorder
type Workorder struct {
	ID            int64
	Number        string
	ProjectID     int64
	FieldID       int64
	LotID         int64
	CropID        int64
	LaborID       int64
	Contractor    string
	Observations  string
	Date          time.Time
	InvestorID    int64
	EffectiveArea decimal.Decimal
	Items         []WorkorderItem

	Base shareddomain.Base
}

// WorkorderItem ...
type WorkorderItem struct {
	SupplyID  int64
	TotalUsed decimal.Decimal
	FinalDose decimal.Decimal
}

// Filtro
type WorkorderFilter struct {
	ProjectID  *int64
	FieldID    *int64
	CustomerID *int64
	CampaignID *int64
}
