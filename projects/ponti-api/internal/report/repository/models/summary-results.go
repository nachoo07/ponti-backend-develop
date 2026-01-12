// Package models contiene los modelos de base de datos para los reportes
package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/db"
	"github.com/shopspring/decimal"
)

// ===== MODELOS DE BASE DE DATOS =====

// SummaryResultsModel representa el modelo de base de datos para resumen de resultados por cultivo
type SummaryResultsModel struct {
	ProjectID int64  `gorm:"column:project_id"`
	CropID    int64  `gorm:"column:current_crop_id"`
	CropName  string `gorm:"column:crop_name"`

	// Métricas por cultivo
	SurfaceHa          decimal.Decimal `gorm:"column:surface_ha"`
	NetIncomeUsd       decimal.Decimal `gorm:"column:net_income_usd"`
	DirectCostsUsd     decimal.Decimal `gorm:"column:direct_costs_usd"`
	RentUsd            decimal.Decimal `gorm:"column:rent_usd"`
	StructureUsd       decimal.Decimal `gorm:"column:structure_usd"`
	TotalInvestedUsd   decimal.Decimal `gorm:"column:total_invested_usd"`
	OperatingResultUsd decimal.Decimal `gorm:"column:operating_result_usd"`
	CropReturnPct      decimal.Decimal `gorm:"column:crop_return_pct"`

	// Totales del proyecto (para comparación)
	TotalSurfaceHa          decimal.Decimal `gorm:"column:total_surface_ha"`
	TotalNetIncomeUsd       decimal.Decimal `gorm:"column:total_net_income_usd"`
	TotalDirectCostsUsd     decimal.Decimal `gorm:"column:total_direct_costs_usd"`
	TotalRentUsd            decimal.Decimal `gorm:"column:total_rent_usd"`
	TotalStructureUsd       decimal.Decimal `gorm:"column:total_structure_usd"`
	TotalInvestedProjectUsd decimal.Decimal `gorm:"column:total_invested_project_usd"`
	TotalOperatingResultUsd decimal.Decimal `gorm:"column:total_operating_result_usd"`
	ProjectReturnPct        decimal.Decimal `gorm:"column:project_return_pct"`
}

// TableName especifica el nombre de la tabla para GORM
// Usa v4_report.summary_results si REPORT_SCHEMA=v4_report, sino v3_report_summary_results_view
func (SummaryResultsModel) TableName() string {
	return db.SummaryView()
}

// ===== MAPPERS =====

// ToDomainSummaryResults convierte de modelo a dominio
func (m *SummaryResultsModel) ToDomainSummaryResults() *domain.SummaryResults {
	return &domain.SummaryResults{
		ProjectID:               m.ProjectID,
		CropID:                  m.CropID,
		CropName:                m.CropName,
		SurfaceHa:               m.SurfaceHa,
		NetIncomeUsd:            m.NetIncomeUsd,
		DirectCostsUsd:          m.DirectCostsUsd,
		RentUsd:                 m.RentUsd,
		StructureUsd:            m.StructureUsd,
		TotalInvestedUsd:        m.TotalInvestedUsd,
		OperatingResultUsd:      m.OperatingResultUsd,
		CropReturnPct:           m.CropReturnPct,
		TotalSurfaceHa:          m.TotalSurfaceHa,
		TotalNetIncomeUsd:       m.TotalNetIncomeUsd,
		TotalDirectCostsUsd:     m.TotalDirectCostsUsd,
		TotalRentUsd:            m.TotalRentUsd,
		TotalStructureUsd:       m.TotalStructureUsd,
		TotalInvestedProjectUsd: m.TotalInvestedProjectUsd,
		TotalOperatingResultUsd: m.TotalOperatingResultUsd,
		ProjectReturnPct:        m.ProjectReturnPct,
	}
}

// ===== FUNCIONES DE AGRUPACIÓN =====

// GroupSummaryResultsByCrop agrupa los resultados por cultivo
func GroupSummaryResultsByCrop(results []*domain.SummaryResults) map[string][]domain.SummaryResults {
	grouped := make(map[string][]domain.SummaryResults)

	for _, result := range results {
		cropName := result.CropName
		if cropName == "" {
			cropName = "Sin cultivo"
		}
		grouped[cropName] = append(grouped[cropName], *result)
	}

	return grouped
}

// CalculateProjectTotals calcula los totales del proyecto
func CalculateProjectTotals(results []*domain.SummaryResults) *domain.ProjectTotals {
	if len(results) == 0 {
		return &domain.ProjectTotals{}
	}

	// Usar los totales del primer resultado (ya vienen calculados de la vista)
	first := results[0]
	return &domain.ProjectTotals{
		TotalSurfaceHa:          first.TotalSurfaceHa,
		TotalNetIncomeUsd:       first.TotalNetIncomeUsd,
		TotalDirectCostsUsd:     first.TotalDirectCostsUsd,
		TotalRentUsd:            first.TotalRentUsd,
		TotalStructureUsd:       first.TotalStructureUsd,
		TotalInvestedProjectUsd: first.TotalInvestedProjectUsd,
		TotalOperatingResultUsd: first.TotalOperatingResultUsd,
		ProjectReturnPct:        first.ProjectReturnPct,
	}
}
