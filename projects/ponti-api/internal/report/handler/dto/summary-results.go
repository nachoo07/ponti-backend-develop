// Package dto holds the Data Transfer Objects for reports.
package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// Decimal0 es un wrapper de decimal.Decimal que serializa sin decimales (redondeo al entero más cercano)
type Decimal0 struct {
	decimal.Decimal
}

// MarshalJSON implementa json.Marshaler para formatear sin decimales
func (d Decimal0) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Decimal.Round(0).String())
}

// NewDecimal0 crea un Decimal0 desde un decimal.Decimal
func NewDecimal0(d decimal.Decimal) Decimal0 {
	return Decimal0{Decimal: d}
}

// Decimal2 es un wrapper de decimal.Decimal que serializa con 2 decimales
type Decimal2 struct {
	decimal.Decimal
}

// MarshalJSON implementa json.Marshaler para formatear con 2 decimales
func (d Decimal2) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Decimal.Round(2).String())
}

// NewDecimal2 crea un Decimal2 desde un decimal.Decimal
func NewDecimal2(d decimal.Decimal) Decimal2 {
	return Decimal2{Decimal: d}
}

// Decimal3 es un wrapper de decimal.Decimal que serializa con 3 decimales
// NOTA: Se mantiene para compatibilidad con código existente que ya tiene Round(3) configurado
type Decimal3 struct {
	decimal.Decimal
}

// MarshalJSON implementa json.Marshaler para formatear con 3 decimales
func (d Decimal3) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Decimal.Round(3).String())
}

// NewDecimal3 crea un Decimal3 desde un decimal.Decimal
func NewDecimal3(d decimal.Decimal) Decimal3 {
	return Decimal3{Decimal: d}
}

/* =========================
   REQUEST DTOs
========================= */

// SummaryResultsRequest represents the request for summary results report.
type SummaryResultsRequest struct {
	ProjectID  *int64 `form:"project_id" binding:"omitempty"`
	CustomerID *int64 `form:"customer_id" binding:"omitempty"`
	CampaignID *int64 `form:"campaign_id" binding:"omitempty"`
	FieldID    *int64 `form:"field_id" binding:"omitempty"`
}

/* =========================
   RESPONSE DTOs
========================= */

// SummaryResultsResponse represents the summary results report response.
type SummaryResultsResponse struct {
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`

	// Resultados por cultivo
	Crops []CropSummaryResponse `json:"crops"`

	// Totales del proyecto (GRAL CAMPOS)
	Totals ProjectTotalsResponse `json:"totals"`

	// Resumen general de cultivos (GRAL CULTIVOS)
	GeneralCrops GeneralCropsResponse `json:"general_crops"`
}

// CropSummaryResponse represents a crop summary in the report.
type CropSummaryResponse struct {
	CropID   int64  `json:"crop_id"`
	CropName string `json:"crop_name"`

	// Métricas por cultivo (sin decimales según reglas de formato)
	SurfaceHa          Decimal0 `json:"surface_ha"`
	NetIncomeUsd       Decimal0 `json:"net_income_usd"`
	DirectCostsUsd     Decimal0 `json:"direct_costs_usd"`
	RentUsd            Decimal0 `json:"rent_usd"`
	StructureUsd       Decimal0 `json:"structure_usd"`
	TotalInvestedUsd   Decimal0 `json:"total_invested_usd"`
	OperatingResultUsd Decimal0 `json:"operating_result_usd"`
	CropReturnPct      Decimal0 `json:"crop_return_pct"`
}

// ProjectTotalsResponse represents the project totals in the report (GRAL CAMPOS).
type ProjectTotalsResponse struct {
	TotalSurfaceHa          Decimal0 `json:"total_surface_ha"`
	TotalNetIncomeUsd       Decimal0 `json:"total_net_income_usd"`
	TotalDirectCostsUsd     Decimal0 `json:"total_direct_costs_usd"`
	TotalRentUsd            Decimal0 `json:"total_rent_usd"`
	TotalStructureUsd       Decimal0 `json:"total_structure_usd"`
	TotalInvestedProjectUsd Decimal0 `json:"total_invested_project_usd"`
	TotalOperatingResultUsd Decimal0 `json:"total_operating_result_usd"`
	ProjectReturnPct        Decimal0 `json:"project_return_pct"`
}

// GeneralCropsResponse represents the general crops summary (GRAL CULTIVOS).
type GeneralCropsResponse struct {
	TotalSurfaceHa          Decimal0 `json:"total_surface_ha"`
	TotalNetIncomeUsd       Decimal0 `json:"total_net_income_usd"`
	TotalDirectCostsUsd     Decimal0 `json:"total_direct_costs_usd"`
	TotalRentUsd            Decimal0 `json:"total_rent_usd"`
	TotalStructureUsd       Decimal0 `json:"total_structure_usd"`
	TotalInvestedProjectUsd Decimal0 `json:"total_invested_project_usd"`
	TotalOperatingResultUsd Decimal0 `json:"total_operating_result_usd"`
	ProjectReturnPct        Decimal0 `json:"project_return_pct"`
}

/* =========================
   MAPPING FUNCTIONS
========================= */

// ToDomainSummaryResultsFilter maps DTO to domain filters.
func ToDomainSummaryResultsFilter(in SummaryResultsRequest) domain.SummaryResultsFilter {
	return domain.SummaryResultsFilter{
		ProjectID:  in.ProjectID,
		CustomerID: in.CustomerID,
		CampaignID: in.CampaignID,
		FieldID:    in.FieldID,
	}
}

// FromDomainSummaryResults maps domain to DTO response.
func FromDomainSummaryResults(d *domain.SummaryResultsResponse) *SummaryResultsResponse {
	// Mapear cultivos (sin decimales según reglas de formato)
	crops := make([]CropSummaryResponse, len(d.Crops))
	for i, crop := range d.Crops {
		crops[i] = CropSummaryResponse{
			CropID:             crop.CropID,
			CropName:           crop.CropName,
			SurfaceHa:          NewDecimal0(crop.SurfaceHa),
			NetIncomeUsd:       NewDecimal0(crop.NetIncomeUsd),
			DirectCostsUsd:     NewDecimal0(crop.DirectCostsUsd),
			RentUsd:            NewDecimal0(crop.RentUsd),
			StructureUsd:       NewDecimal0(crop.StructureUsd),
			TotalInvestedUsd:   NewDecimal0(crop.TotalInvestedUsd),
			OperatingResultUsd: NewDecimal0(crop.OperatingResultUsd),
			CropReturnPct:      NewDecimal0(crop.CropReturnPct),
		}
	}

	// Mapear totales (sin decimales según reglas de formato)
	totals := ProjectTotalsResponse{
		TotalSurfaceHa:          NewDecimal0(d.Totals.TotalSurfaceHa),
		TotalNetIncomeUsd:       NewDecimal0(d.Totals.TotalNetIncomeUsd),
		TotalDirectCostsUsd:     NewDecimal0(d.Totals.TotalDirectCostsUsd),
		TotalRentUsd:            NewDecimal0(d.Totals.TotalRentUsd),
		TotalStructureUsd:       NewDecimal0(d.Totals.TotalStructureUsd),
		TotalInvestedProjectUsd: NewDecimal0(d.Totals.TotalInvestedProjectUsd),
		TotalOperatingResultUsd: NewDecimal0(d.Totals.TotalOperatingResultUsd),
		ProjectReturnPct:        NewDecimal0(d.Totals.ProjectReturnPct),
	}

	// Mapear cultivos generales (GRAL CULTIVOS) sin decimales según reglas de formato
	generalCrops := GeneralCropsResponse{
		TotalSurfaceHa:          NewDecimal0(d.GeneralCrops.TotalSurfaceHa),
		TotalNetIncomeUsd:       NewDecimal0(d.GeneralCrops.TotalNetIncomeUsd),
		TotalDirectCostsUsd:     NewDecimal0(d.GeneralCrops.TotalDirectCostsUsd),
		TotalRentUsd:            NewDecimal0(d.GeneralCrops.TotalRentUsd),
		TotalStructureUsd:       NewDecimal0(d.GeneralCrops.TotalStructureUsd),
		TotalInvestedProjectUsd: NewDecimal0(d.GeneralCrops.TotalInvestedProjectUsd),
		TotalOperatingResultUsd: NewDecimal0(d.GeneralCrops.TotalOperatingResultUsd),
		ProjectReturnPct:        NewDecimal0(d.GeneralCrops.ProjectReturnPct),
	}

	return &SummaryResultsResponse{
		ProjectID:    d.ProjectID,
		ProjectName:  d.ProjectName,
		CustomerID:   d.CustomerID,
		CustomerName: d.CustomerName,
		CampaignID:   d.CampaignID,
		CampaignName: d.CampaignName,
		Crops:        crops,
		Totals:       totals,
		GeneralCrops: generalCrops,
	}
}
