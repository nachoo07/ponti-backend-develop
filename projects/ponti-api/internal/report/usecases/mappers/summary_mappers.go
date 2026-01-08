// Package mappers proporciona funciones de mapeo para los casos de uso de reportes
package mappers

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// SummaryResponseMapper maneja el mapeo de respuestas de resumen
type SummaryResponseMapper struct{}

// NewSummaryResponseMapper crea una nueva instancia del mapper
func NewSummaryResponseMapper() *SummaryResponseMapper {
	return &SummaryResponseMapper{}
}

// BuildEmptyResponse construye una respuesta vacía
func (m *SummaryResponseMapper) BuildEmptyResponse() *domain.SummaryResultsResponse {
	return &domain.SummaryResultsResponse{
		Crops:        []domain.SummaryResults{},
		Totals:       domain.ProjectTotals{},
		GeneralCrops: domain.GeneralCrops{},
	}
}

// BuildResponse construye la respuesta completa con datos
func (m *SummaryResponseMapper) BuildResponse(
	projectInfo *domain.ProjectInfo,
	crops []domain.SummaryResults,
	totals *domain.ProjectTotals,
) *domain.SummaryResultsResponse {
	// GRAL CULTIVOS debe mostrar valores por hectárea (promedios)
	generalCrops := m.calculateGeneralCrops(totals)

	return &domain.SummaryResultsResponse{
		ProjectID:    projectInfo.ProjectID,
		ProjectName:  projectInfo.ProjectName,
		CustomerID:   projectInfo.CustomerID,
		CustomerName: projectInfo.CustomerName,
		CampaignID:   projectInfo.CampaignID,
		CampaignName: projectInfo.CampaignName,
		Crops:        crops,
		Totals:       *totals,
		GeneralCrops: generalCrops,
	}
}

// ConvertToPointers convierte un slice de valores a un slice de punteros
func (m *SummaryResponseMapper) ConvertToPointers(results []domain.SummaryResults) []*domain.SummaryResults {
	pointers := make([]*domain.SummaryResults, len(results))
	for i := range results {
		pointers[i] = &results[i]
	}
	return pointers
}

// calculateGeneralCrops calcula los valores por hectárea para GRAL CULTIVOS
func (m *SummaryResponseMapper) calculateGeneralCrops(totals *domain.ProjectTotals) domain.GeneralCrops {
	return domain.GeneralCrops{
		// Superficie total se mantiene igual
		TotalSurfaceHa: totals.TotalSurfaceHa,

		// Todos los demás valores se dividen por superficie para obtener promedio por hectárea
		TotalNetIncomeUsd:       m.divideByHectares(totals.TotalNetIncomeUsd, totals.TotalSurfaceHa),
		TotalDirectCostsUsd:     m.divideByHectares(totals.TotalDirectCostsUsd, totals.TotalSurfaceHa),
		TotalRentUsd:            m.divideByHectares(totals.TotalRentUsd, totals.TotalSurfaceHa),
		TotalStructureUsd:       m.divideByHectares(totals.TotalStructureUsd, totals.TotalSurfaceHa),
		TotalInvestedProjectUsd: m.divideByHectares(totals.TotalInvestedProjectUsd, totals.TotalSurfaceHa),
		TotalOperatingResultUsd: m.divideByHectares(totals.TotalOperatingResultUsd, totals.TotalSurfaceHa),

		// Porcentaje se mantiene igual (no se divide)
		ProjectReturnPct: totals.ProjectReturnPct,
	}
}

// divideByHectares divide un valor por hectáreas de manera segura
func (m *SummaryResponseMapper) divideByHectares(value, hectares decimal.Decimal) decimal.Decimal {
	if hectares.IsZero() || hectares.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	return value.Div(hectares)
}
