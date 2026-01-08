// Package domain contiene las entidades de dominio para los reportes
package domain

import (
	"github.com/shopspring/decimal"
)

// ===== ENTIDADES DE DOMINIO =====

// SummaryResults representa el resumen de resultados por cultivo
type SummaryResults struct {
	ProjectID int64  `json:"project_id"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`

	// Métricas por cultivo
	SurfaceHa          decimal.Decimal `json:"surface_ha"`
	NetIncomeUsd       decimal.Decimal `json:"net_income_usd"`
	DirectCostsUsd     decimal.Decimal `json:"direct_costs_usd"`
	RentUsd            decimal.Decimal `json:"rent_usd"`
	StructureUsd       decimal.Decimal `json:"structure_usd"`
	TotalInvestedUsd   decimal.Decimal `json:"total_invested_usd"`
	OperatingResultUsd decimal.Decimal `json:"operating_result_usd"`
	CropReturnPct      decimal.Decimal `json:"crop_return_pct"`

	// Totales del proyecto (para comparación)
	TotalSurfaceHa          decimal.Decimal `json:"total_surface_ha"`
	TotalNetIncomeUsd       decimal.Decimal `json:"total_net_income_usd"`
	TotalDirectCostsUsd     decimal.Decimal `json:"total_direct_costs_usd"`
	TotalRentUsd            decimal.Decimal `json:"total_rent_usd"`
	TotalStructureUsd       decimal.Decimal `json:"total_structure_usd"`
	TotalInvestedProjectUsd decimal.Decimal `json:"total_invested_project_usd"`
	TotalOperatingResultUsd decimal.Decimal `json:"total_operating_result_usd"`
	ProjectReturnPct        decimal.Decimal `json:"project_return_pct"`
}

// ProjectTotals representa los totales del proyecto (GRAL CAMPOS).
type ProjectTotals struct {
	TotalSurfaceHa          decimal.Decimal `json:"total_surface_ha"`
	TotalNetIncomeUsd       decimal.Decimal `json:"total_net_income_usd"`
	TotalDirectCostsUsd     decimal.Decimal `json:"total_direct_costs_usd"`
	TotalRentUsd            decimal.Decimal `json:"total_rent_usd"`
	TotalStructureUsd       decimal.Decimal `json:"total_structure_usd"`
	TotalInvestedProjectUsd decimal.Decimal `json:"total_invested_project_usd"`
	TotalOperatingResultUsd decimal.Decimal `json:"total_operating_result_usd"`
	ProjectReturnPct        decimal.Decimal `json:"project_return_pct"`
}

// GeneralCrops representa el resumen general de cultivos (GRAL CULTIVOS).
type GeneralCrops struct {
	TotalSurfaceHa          decimal.Decimal `json:"total_surface_ha"`
	TotalNetIncomeUsd       decimal.Decimal `json:"total_net_income_usd"`
	TotalDirectCostsUsd     decimal.Decimal `json:"total_direct_costs_usd"`
	TotalRentUsd            decimal.Decimal `json:"total_rent_usd"`
	TotalStructureUsd       decimal.Decimal `json:"total_structure_usd"`
	TotalInvestedProjectUsd decimal.Decimal `json:"total_invested_project_usd"`
	TotalOperatingResultUsd decimal.Decimal `json:"total_operating_result_usd"`
	ProjectReturnPct        decimal.Decimal `json:"project_return_pct"`
}

// SummaryResultsResponse representa la respuesta completa del reporte
type SummaryResultsResponse struct {
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`

	// Resultados por cultivo
	Crops []SummaryResults `json:"crops"`

	// Totales del proyecto (GRAL CAMPOS)
	Totals ProjectTotals `json:"totals"`

	// Resumen general de cultivos (GRAL CULTIVOS)
	GeneralCrops GeneralCrops `json:"general_crops"`
}

// ===== FILTROS =====

// SummaryResultsFilter representa los filtros para el reporte de resumen de resultados
type SummaryResultsFilter struct {
	ProjectID  *int64 `json:"project_id"`
	CustomerID *int64 `json:"customer_id"`
	CampaignID *int64 `json:"campaign_id"`
	FieldID    *int64 `json:"field_id"`
}
