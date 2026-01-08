// Package domain contiene los modelos de dominio para los reportes
package domain

import (
	"github.com/shopspring/decimal"
)

// ===== DOMAIN MODELS =====

// ReportFilter representa los filtros para los reportes
type ReportFilter struct {
	CustomerID *int64 `json:"customer_id"`
	ProjectID  *int64 `json:"project_id"`
	CampaignID *int64 `json:"campaign_id"`
	FieldID    *int64 `json:"field_id"`
}

// ProjectInfo representa la información básica de un proyecto
type ProjectInfo struct {
	ProjectID    int64  `json:"project_id" gorm:"column:project_id"`
	ProjectName  string `json:"project_name" gorm:"column:project_name"`
	CustomerID   int64  `json:"customer_id" gorm:"column:customer_id"`
	CustomerName string `json:"customer_name" gorm:"column:customer_name"`
	CampaignID   int64  `json:"campaign_id" gorm:"column:campaign_id"`
	CampaignName string `json:"campaign_name" gorm:"column:campaign_name"`
}

// FieldCropMetric representa una métrica por campo y cultivo
type FieldCropMetric struct {
	ProjectID int64  `json:"project_id"`
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`

	// Información general
	SurfaceHa       decimal.Decimal `json:"surface_ha"`
	ProductionTn    decimal.Decimal `json:"production_tn"`
	SownAreaHa      decimal.Decimal `json:"sown_area_ha"`
	HarvestedAreaHa decimal.Decimal `json:"harvested_area_ha"`

	// Rendimiento
	YieldTnHa decimal.Decimal `json:"yield_tn_ha"`

	// Precios y comercialización
	GrossPriceUsdTn     decimal.Decimal `json:"gross_price_usd_tn"`
	FreightCostUsdTn    decimal.Decimal `json:"freight_cost_usd_tn"`
	CommercialCostUsdTn decimal.Decimal `json:"commercial_cost_usd_tn"`
	NetPriceUsdTn       decimal.Decimal `json:"net_price_usd_tn"`

	// Ingreso neto
	NetIncomeUsd   decimal.Decimal `json:"net_income_usd"`
	NetIncomeUsdHa decimal.Decimal `json:"net_income_usd_ha"`

	// Costos directos
	LaborCostsUsd       decimal.Decimal `json:"labor_costs_usd"`
	LaborCostsUsdHa     decimal.Decimal `json:"labor_costs_usd_ha"` // TODO: Campo agregado para vista v3
	SupplyCostsUsd      decimal.Decimal `json:"supply_costs_usd"`
	SupplyCostsUsdHa    decimal.Decimal `json:"supply_costs_usd_ha"` // TODO: Campo agregado para vista v3
	TotalDirectCostsUsd decimal.Decimal `json:"total_direct_costs_usd"`
	DirectCostsUsdHa    decimal.Decimal `json:"direct_costs_usd_ha"`

	// Margen bruto
	GrossMarginUsd   decimal.Decimal `json:"gross_margin_usd"`
	GrossMarginUsdHa decimal.Decimal `json:"gross_margin_usd_ha"`

	// Arriendo
	RentUsd   decimal.Decimal `json:"rent_usd"`
	RentUsdHa decimal.Decimal `json:"rent_usd_ha"`

	// Costos administrativos
	AdministrationUsd   decimal.Decimal `json:"administration_usd"`
	AdministrationUsdHa decimal.Decimal `json:"administration_usd_ha"`

	// Resultado operativo
	OperatingResultUsd   decimal.Decimal `json:"operating_result_usd"`
	OperatingResultUsdHa decimal.Decimal `json:"operating_result_usd_ha"`

	// Total invertido
	TotalInvestedUsd   decimal.Decimal `json:"total_invested_usd"`
	TotalInvestedUsdHa decimal.Decimal `json:"total_invested_usd_ha"`

	// Métricas calculadas
	ReturnPct              decimal.Decimal `json:"return_pct"`
	IndifferenceYieldUsdTn decimal.Decimal `json:"indifference_yield_usd_tn"`
}

// ===== TABLE DOMAIN MODELS =====

// FieldCrop representa la tabla de reporte field-crop
type FieldCrop struct {
	ProjectID    int64             `json:"project_id"`
	ProjectName  string            `json:"project_name"`
	CustomerID   int64             `json:"customer_id"`
	CustomerName string            `json:"customer_name"`
	CampaignID   int64             `json:"campaign_id"`
	CampaignName string            `json:"campaign_name"`
	Columns      []FieldCropColumn `json:"columns"`
	Rows         []FieldCropRow    `json:"rows"`
}

// FieldCropColumn representa una columna en la tabla
type FieldCropColumn struct {
	ID        string `json:"id"` // field_id-crop_id
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`
}

// FieldCropRow representa una fila en la tabla
type FieldCropRow struct {
	Key       string                    `json:"key"`
	Unit      string                    `json:"unit"`
	ValueType string                    `json:"value_type"` // "number" or "text"
	Values    map[string]FieldCropValue `json:"values"`
}

// FieldCropValue representa un valor en la tabla
type FieldCropValue struct {
	Number decimal.Decimal `json:"number"`
}

// LaborMetric representa una métrica de labor
type LaborMetric struct {
	LaborID        int64           `json:"labor_id"`
	LaborName      string          `json:"labor_name"`
	CategoryID     int64           `json:"category_id"`
	CategoryName   string          `json:"category_name"`
	SurfaceHa      decimal.Decimal `json:"surface_ha"`
	CostUsd        decimal.Decimal `json:"cost_usd"`
	CostPerHa      decimal.Decimal `json:"cost_per_ha"`
	WorkOrderCount int64           `json:"workorder_count"`
}

// SupplyMetric representa una métrica de supply
type SupplyMetric struct {
	SupplyID       int64           `json:"supply_id"`
	SupplyName     string          `json:"supply_name"`
	CategoryID     int64           `json:"category_id"`
	CategoryName   string          `json:"category_name"`
	SurfaceHa      decimal.Decimal `json:"surface_ha"`
	QuantityUsed   decimal.Decimal `json:"quantity_used"`
	CostUsd        decimal.Decimal `json:"cost_usd"`
	CostPerHa      decimal.Decimal `json:"cost_per_ha"`
	WorkOrderCount int64           `json:"workorder_count"`
}
