// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/mappers"
)

// ===== PORTS (Hexagonal Architecture) =====

// ReportRepositoryPort define la interfaz del repositorio (Puerto de salida)
type ReportRepositoryPort interface {
	GetFieldCropMetrics(domain.ReportFilter) ([]domain.FieldCropMetric, error)
	GetProjectInfo(int64) (*domain.ProjectInfo, error)
	BuildFieldCrop(domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(context.Context, domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResults(domain.SummaryResultsFilter) ([]domain.SummaryResults, error)
}

// ReportUseCasePort define la interfaz del caso de uso (Puerto de entrada)
type ReportUseCasePort interface {
	GetFieldCropReport(domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(context.Context, domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResultsReport(domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error)
}

// ===== USE CASE IMPLEMENTATION =====

// ReportUseCase implementa la lógica de negocio para reportes
type ReportUseCase struct {
	repository    ReportRepositoryPort
	validator     *usecases.ReportFilterValidator
	summaryMapper *mappers.SummaryResponseMapper
}

// NewReportUseCase crea una nueva instancia del caso de uso
func NewReportUseCase(repository ReportRepositoryPort) *ReportUseCase {
	return &ReportUseCase{
		repository:    repository,
		validator:     usecases.NewReportFilterValidator(),
		summaryMapper: mappers.NewSummaryResponseMapper(),
	}
}

// ===== REPORTE POR CAMPO/CULTIVO =====

// GetFieldCropReport obtiene el reporte por campo/cultivo
func (uc *ReportUseCase) GetFieldCropReport(filters domain.ReportFilter) (*domain.FieldCrop, error) {

	// Obtener reporte del repositorio
	report, err := uc.repository.BuildFieldCrop(filters)
	if err != nil {
		return nil, fmt.Errorf("error al obtener reporte de campo/cultivo: %w", err)
	}

	return report, nil
}

// GetInvestorContributionReport obtiene el reporte de aportes de inversores
func (uc *ReportUseCase) GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error) {
	// Todos los filtros son opcionales - no hay validaciones requeridas

	// Obtener datos desde el repository (que consulta la vista de la DB)
	report, err := uc.repository.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo reporte de aportes de inversores: %w", err)
	}

	return report, nil
}

// ===== REPORTE DE RESUMEN DE RESULTADOS =====

// GetSummaryResultsReport obtiene el reporte de resumen de resultados
func (uc *ReportUseCase) GetSummaryResultsReport(filters domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error) {
	// Validar que al menos un filtro esté presente
	if err := uc.validator.ValidateAtLeastOneFilter(filters); err != nil {
		return nil, err
	}

	// Obtener datos del repositorio
	results, err := uc.repository.GetSummaryResults(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen de resultados: %w", err)
	}

	// Retornar respuesta vacía si no hay resultados
	if len(results) == 0 {
		return uc.buildEmptySummaryResponse(), nil
	}

	// Construir respuesta con datos
	return uc.buildSummaryResponse(results)
}

// ===== FUNCIONES PRIVADAS (DRY) =====

// buildEmptySummaryResponse construye una respuesta vacía usando el mapper
func (uc *ReportUseCase) buildEmptySummaryResponse() *domain.SummaryResultsResponse {
	return uc.summaryMapper.BuildEmptyResponse()
}

// buildSummaryResponse construye la respuesta completa con datos usando el mapper
func (uc *ReportUseCase) buildSummaryResponse(results []domain.SummaryResults) (*domain.SummaryResultsResponse, error) {
	// Obtener información del proyecto del primer resultado
	projectInfo, err := uc.repository.GetProjectInfo(results[0].ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project information: %w", err)
	}

	// Calcular totales del proyecto
	totales := uc.calculateProjectTotals(results)

	// Usar el mapper para construir la respuesta
	return uc.summaryMapper.BuildResponse(projectInfo, results, totales), nil
}

// calculateProjectTotals calcula los totales del proyecto
func (uc *ReportUseCase) calculateProjectTotals(results []domain.SummaryResults) *domain.ProjectTotals {
	// Usar el mapper para convertir a punteros
	resultsPtr := uc.summaryMapper.ConvertToPointers(results)
	return models.CalculateProjectTotals(resultsPtr)
}
