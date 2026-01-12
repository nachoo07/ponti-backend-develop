// Package models contiene los modelos de datos para el módulo de labor
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// LaborListItem - Modelo que incluye campos calculados de la vista v3_labor_list
type LaborListItem struct {
	WorkorderID        int64           `gorm:"column:workorder_id"`
	WorkorderNumber    string          `gorm:"column:workorder_number"`
	Date               time.Time       `gorm:"column:date"`
	ProjectID          int64           `gorm:"column:project_id"`
	ProjectName        string          `gorm:"column:project_name"`
	FieldID            int64           `gorm:"column:field_id"`
	FieldName          string          `gorm:"column:field_name"`
	LotID              *int64          `gorm:"column:lot_id"`
	LotName            *string         `gorm:"column:lot_name"`
	CropID             *int64          `gorm:"column:crop_id"`
	CropName           *string         `gorm:"column:crop_name"`
	LaborID            int64           `gorm:"column:labor_id"`
	LaborName          string          `gorm:"column:labor_name"`
	LaborCategoryID    *int64          `gorm:"column:labor_category_id"`
	LaborCategoryName  *string         `gorm:"column:labor_category_name"`
	Contractor         string          `gorm:"column:contractor"`
	ContractorName     string          `gorm:"column:contractor_name"`
	SurfaceHa          decimal.Decimal `gorm:"column:surface_ha"`
	CostPerHa          decimal.Decimal `gorm:"column:cost_per_ha"`
	TotalLaborCost     decimal.Decimal `gorm:"column:total_labor_cost"`
	DollarAverageMonth decimal.Decimal `gorm:"column:dollar_average_month"`
	USDAvgValue        decimal.Decimal `gorm:"column:usd_avg_value"`
	USDCostHa          decimal.Decimal `gorm:"column:usd_cost_ha"`
	USDNetTotal        decimal.Decimal `gorm:"column:usd_net_total"`
	InvestorID         *int64          `gorm:"column:investor_id"`
	InvestorName       *string         `gorm:"column:investor_name"`

	// Campos adicionales de joins
	InvoiceID      *int64     `gorm:"column:invoice_id"`
	InvoiceNumber  *string    `gorm:"column:invoice_number"`
	InvoiceCompany *string    `gorm:"column:invoice_company"`
	InvoiceDate    *time.Time `gorm:"column:invoice_date"`
	InvoiceStatus  *string    `gorm:"column:invoice_status"`
}
