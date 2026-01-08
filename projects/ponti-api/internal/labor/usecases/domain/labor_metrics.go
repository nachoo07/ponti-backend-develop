package domain

import "github.com/shopspring/decimal"

// LaborMetrics agrega superficie y costo promedio por hectárea para labors
type LaborMetrics struct {
	SurfaceHa    decimal.Decimal
	NetTotalCost decimal.Decimal
	AvgCostPerHa decimal.Decimal
}

// LaborFilter permite filtrado opcional por proyecto y campo
type LaborFilter struct {
	ProjectID *int64
	FieldID   *int64
}
