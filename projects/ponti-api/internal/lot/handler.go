// Package lot expone los handlers HTTP para la entidad Lot.
package lot

import (
	// standard library
	"context"
	"net/http"
	"strconv"

	// third-party
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	// pkg
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	// excel
	lotExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/excel"

	// project
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type UseCasesPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	UpdateLotTons(context.Context, int64, decimal.Decimal) error
	DeleteLot(context.Context, int64) error
	ListLotsByField(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	GetMetrics(context.Context, int64, int64, int64) (*domain.LotMetrics, error)
	ListLots(context.Context, int64, int64, int64, int, int) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error)
	ExportLots(context.Context, int64, int64, int64, int, int) ([]byte, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/lots"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", ValidateLotRequest(), h.CreateLot)
		public.GET("", h.ListLots)
		public.GET("/metrics", h.GetMetrics)
		public.PUT("/:id/tons", ValidateLotTonsUpdate(), h.UpdateLotTons)
		public.GET("/:id", h.GetLot)
		public.PUT("/:id", ValidateLotUpdate(), h.UpdateLot)
		public.DELETE("/:id", h.DeleteLot)
		public.GET("/export", h.ExportLots)
	}
}

func (h *Handler) CreateLot(c *gin.Context) {
	// El lote ya fue validado por el middleware ValidateLotRequest
	req := c.MustGet("validated_lot").(*dto.Lot)

	lotDomain, err := req.ToDomain()
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrValidation, "invalid domain conversion", err))
		return
	}

	newID, err := h.ucs.CreateLot(c.Request.Context(), lotDomain)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.CreateLotResponse{Message: "Lot created successfully", ID: newID})
}

func (h *Handler) ListLots(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	fieldID, _ := strconv.ParseInt(c.Query("field_id"), 10, 64)
	cropID, _ := strconv.ParseInt(c.Query("crop_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "1000"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 1000
	}
	// Los filtros por ID son opcionales para permitir búsquedas globales
	// Si no se proporcionan filtros, se retornan todos los lotes
	// Cap de paginación
	if pageSize > 1000 {
		pageSize = 1000
	}

	rows, total, sumSowed, sumCost, err := h.ucs.ListLots(c.Request.Context(), projectID, fieldID, cropID, page, pageSize)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	pageInfo := types.NewPageInfo(page, pageSize, int64(total))
	resp := dto.FromDomainList(pageInfo, rows, sumSowed, sumCost)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid lot id", err))
		return
	}
	lot, err := h.ucs.GetLot(c.Request.Context(), id)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(lot))
}

func (h *Handler) UpdateLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid lot id", err))
		return
	}

	// El lote ya fue validado por el middleware ValidateLotUpdate
	req := c.MustGet("validated_lot").(*dto.LotUpdate)

	dom, err := req.ToDomain()
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrValidation, "invalid domain conversion", err))
		return
	}
	dom.ID = id

	// Obtener user ID del contexto para campos de auditoría
	if userID, err := sharedmodels.ConvertStringToID(c.Request.Context()); err == nil {
		dom.Base.UpdatedBy = &userID
	}

	// Si el cliente no envía field_id, usamos el existente para evitar inconsistencias
	if dom.FieldID == 0 {
		if cur, getErr := h.ucs.GetLot(c.Request.Context(), id); getErr == nil {
			dom.FieldID = cur.FieldID
		} else {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "field_id is required"})
			return
		}
	}
	if err := h.ucs.UpdateLot(c.Request.Context(), dom); err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateLotTons(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid lot id", err))
		return
	}

	// Las toneladas ya fueron validadas por el middleware ValidateLotTonsUpdate
	tons := c.MustGet("validated_tons").(decimal.Decimal)

	if err := h.ucs.UpdateLotTons(c.Request.Context(), id, tons); err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		types.NewErrorResponseHelper().InvalidPayload(c, types.NewError(types.ErrInvalidID, "invalid lot id", err))
		return
	}
	if err := h.ucs.DeleteLot(c.Request.Context(), id); err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetMetrics(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	fieldID, _ := strconv.ParseInt(c.Query("field_id"), 10, 64)
	cropID, _ := strconv.ParseInt(c.Query("crop_id"), 10, 64)

	// Los filtros por ID son opcionales para permitir búsquedas globales
	// Si no se proporcionan filtros, se retornan métricas de todos los lotes

	m, err := h.ucs.GetMetrics(c.Request.Context(), projectID, fieldID, cropID)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportLots(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	fieldID, _ := strconv.ParseInt(c.Query("field_id"), 10, 64)
	cropID, _ := strconv.ParseInt(c.Query("crop_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "1000"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 1000
	}
	// Los filtros por ID son opcionales para permitir búsquedas globales
	// Si no se proporcionan filtros, se retornan todos los lotes
	// Cap de paginación
	if pageSize > 1000 {
		pageSize = 1000
	}

	data, err := h.ucs.ExportLots(c.Request.Context(), projectID, fieldID, cropID, page, pageSize)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := lotExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
