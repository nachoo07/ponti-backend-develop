// Package data_integrity proporciona funcionalidades para validar la integridad de datos
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
//
// ESTE MÓDULO CONTIENE CÁLCULOS CRÍTICOS DE INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
package data_integrity

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/data-integrity/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/data-integrity/usecases/domain"
)

// UseCasesPort define la interfaz para los casos de uso
type UseCasesPort interface {
	CheckCostsCoherence(ctx context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error)
}

// GinEnginePort define la interfaz para el servidor Gin
type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

// ConfigAPIPort define la interfaz para la configuración de la API
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

// Handler maneja las peticiones HTTP del módulo data_integrity
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler crea una nueva instancia del handler
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

// Routes registra las rutas del módulo
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	base := h.acf.APIBaseURL() + "/data-integrity"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(base)
	{
		public.GET("/costs-check", h.CheckCostsCoherence)
	}
}

// CheckCostsCoherence valida la coherencia de costos directos totales
// @Summary Validar coherencia de costos directos
// @Description Valida que el valor de costos directos totales sea consistente en todos los módulos
// @Tags data-integrity
// @Accept json
// @Produce json
// @Param project_id query int false "Project ID"
// @Success 200 {object} dto.IntegrityReportResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /data-integrity/costs-check [get]
func (h *Handler) CheckCostsCoherence(c *gin.Context) {
	// Parsear query params
	var filter domain.CostsCheckFilter

	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project_id"})
			return
		}
		filter.ProjectID = &projectID
	}

	// Ejecutar caso de uso
	report, err := h.ucs.CheckCostsCoherence(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	// Convertir a DTO y retornar
	response := dto.ToIntegrityReportResponse(report)
	c.JSON(http.StatusOK, response)
}
