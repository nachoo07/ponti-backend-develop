// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// UseCasesPort define la interfaz para los casos de uso
type UseCasesPort interface {
	GetFieldCropReport(domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(context.Context, domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResultsReport(domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error)
}

// GinEnginePort define la interfaz para el motor Gin
type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(context.Context) error
}

// ConfigAPIPort define la interfaz para la configuración de API
type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

// MiddlewaresEnginePort define la interfaz para los middlewares
type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

// ReportHandler maneja las peticiones HTTP para reportes
type ReportHandler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewReportHandler crea una nueva instancia del handler
func NewReportHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *ReportHandler {
	return &ReportHandler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// ===== RUTAS =====

// Routes registra todas las rutas del handler
func (h *ReportHandler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/reports"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	reports := r.Group(baseURL)
	{
		// Handler genérico para todos los reportes
		reports.GET("/:type", h.GetReport)
	}
}

// ===== HANDLER GENÉRICO =====

// GetReport maneja todas las peticiones de reportes de forma unificada
func (h *ReportHandler) GetReport(c *gin.Context) {
	reportType := c.Param("type")

	// Validar tipo de reporte
	if !h.isValidReportType(reportType) {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid report type", "Valid types: field-crop, investor-contribution, summary-results")
		return
	}

	// Parsear filtros según el tipo de reporte
	filters, err := h.parseFiltersByType(c, reportType)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid filters", err.Error())
		return
	}

	// Obtener reporte según el tipo
	report, err := h.buildReportByType(c, reportType, filters)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	h.sendSuccessResponse(c, report)
}

// ===== FUNCIONES AUXILIARES =====

// isValidReportType valida si el tipo de reporte es válido
func (h *ReportHandler) isValidReportType(reportType string) bool {
	validTypes := map[string]bool{
		"field-crop":            true,
		"investor-contribution": true,
		"summary-results":       true,
	}
	return validTypes[reportType]
}

// parseFiltersByType parsea los filtros según el tipo de reporte
func (h *ReportHandler) parseFiltersByType(c *gin.Context, reportType string) (interface{}, error) {
	switch reportType {
	case "field-crop", "investor-contribution":
		return h.parseReportFilters(c)
	case "summary-results":
		return h.parseSummaryFilters(c)
	default:
		return nil, nil
	}
}

// parseReportFilters parsea filtros para reportes estándar
func (h *ReportHandler) parseReportFilters(c *gin.Context) (domain.ReportFilter, error) {
	filters := domain.ReportFilter{}

	// Parsear customer_id
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.CustomerID = &customerID
	}

	// Parsear project_id
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.ProjectID = &projectID
	}

	// Parsear campaign_id
	if campaignIDStr := c.Query("campaign_id"); campaignIDStr != "" {
		campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.CampaignID = &campaignID
	}

	// Parsear field_id
	if fieldIDStr := c.Query("field_id"); fieldIDStr != "" {
		fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.FieldID = &fieldID
	}

	return filters, nil
}

// parseSummaryFilters parsea filtros para reporte de resumen
func (h *ReportHandler) parseSummaryFilters(c *gin.Context) (domain.SummaryResultsFilter, error) {
	var request dto.SummaryResultsRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		return domain.SummaryResultsFilter{}, err
	}
	return dto.ToDomainSummaryResultsFilter(request), nil
}

// buildReportByType construye el reporte según el tipo
func (h *ReportHandler) buildReportByType(c *gin.Context, reportType string, filters interface{}) (interface{}, error) {
	switch reportType {
	case "field-crop":
		report, err := h.ucs.GetFieldCropReport(filters.(domain.ReportFilter))
		if err != nil {
			return nil, err
		}
		return dto.BuildFieldCropResponse(report), nil

	case "investor-contribution":
		report, err := h.ucs.GetInvestorContributionReport(c.Request.Context(), filters.(domain.ReportFilter))
		if err != nil {
			return nil, err
		}
		return dto.FromDomainInvestorReport(report), nil

	case "summary-results":
		report, err := h.ucs.GetSummaryResultsReport(filters.(domain.SummaryResultsFilter))
		if err != nil {
			return nil, err
		}
		return dto.FromDomainSummaryResults(report), nil

	default:
		return nil, nil
	}
}

// sendErrorResponse envía una respuesta de error estandarizada
func (h *ReportHandler) sendErrorResponse(c *gin.Context, statusCode int, message string, details string) {
	c.JSON(statusCode, gin.H{
		"error":   message,
		"details": details,
	})
}

// sendSuccessResponse envía una respuesta exitosa
func (h *ReportHandler) sendSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
