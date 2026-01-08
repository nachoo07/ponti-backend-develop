// Package domain define estructuras de dominio para lot.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type LotTable struct {
	ID             int64
	ProjectID      int64
	FieldID        int64
	ProjectName    string
	FieldName      string
	LotName        string
	PreviousCrop   string
	PreviousCropID int64
	CurrentCrop    string
	CurrentCropID  int64
	Variety        string
	SowedArea      decimal.Decimal
	Hectares       decimal.Decimal
	Season         string
	Tons           decimal.Decimal
	Dates          []LotDates
	UpdatedAt      *time.Time

	AdminCost            decimal.Decimal
	HarvestedArea        decimal.Decimal
	HarvestDate          *time.Time
	CostUsdPerHa         decimal.Decimal
	YieldTnPerHa         decimal.Decimal
	IncomeNetPerHa       decimal.Decimal
	RentPerHa            decimal.Decimal
	ActiveTotalPerHa     decimal.Decimal
	OperatingResultPerHa decimal.Decimal
}

type LotDates struct {
	SowingDate  *time.Time
	HarvestDate *time.Time
	Sequence    int
}
