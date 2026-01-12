package workorder

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	workorderexcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/excel"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type UseCasesPort interface {
	CreateWorkorder(context.Context, *domain.Workorder) (int64, error)
	GetWorkorderByID(context.Context, int64) (*domain.Workorder, error)
	DuplicateWorkorder(context.Context, string) (string, error)
	UpdateWorkorderByID(context.Context, *domain.Workorder) error
	DeleteWorkorderByID(context.Context, int64) error
	ListWorkorders(context.Context, domain.WorkorderFilter, types.Input) ([]domain.WorkorderListElement, types.PageInfo, error)
	GetMetrics(context.Context, domain.WorkorderFilter) (*domain.WorkorderMetrics, error)
	ExportWorkorders(context.Context, domain.WorkorderFilter, types.Input) ([]byte, error)
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
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
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	base := h.acf.APIBaseURL() + "/workorders"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	grp := r.Group(base)
	{

		grp.POST("", h.CreateWorkorder)
		grp.GET("/:id", h.GetWorkorderByID)
		grp.PUT("/:id", h.UpdateWorkorderByID)
		grp.DELETE("/:id", h.DeleteWorkorderByID)
		grp.POST("/:number/duplicate", h.DuplicateWorkorder)
		grp.GET("", h.ListWorkorders)
		grp.GET("/metrics", h.GetMetrics)
		grp.GET("/export", h.ExportWorkorders)
	}
}

func (h *Handler) CreateWorkorder(c *gin.Context) {
	var req dto.Workorder
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	id, err := h.ucs.CreateWorkorder(c, req.ToDomain())
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, dto.WorkorderResponse{
		Message: "Workorder created",
		Number:  id,
	})
}

func (h *Handler) GetWorkorderByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	wo, err := h.ucs.GetWorkorderByID(c, id)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(wo))
}

func (h *Handler) DuplicateWorkorder(c *gin.Context) {
	// orig := c.Param("number")
	// newNum, err := h.ucs.DuplicateWorkorder(c.Request.Context(), orig)
	// if err != nil {
	// 	apiErr, status := types.NewAPIError(err)
	// 	c.JSON(status, apiErr.ToResponse())
	// 	return
	// }
	c.JSON(http.StatusCreated, dto.WorkorderResponse{
		Message: "Workorder duplicated",
		Number:  0,
	})
}

func (h *Handler) UpdateWorkorderByID(c *gin.Context) {
	var req dto.Workorder
	if err := c.ShouldBindJSON(&req); err != nil {
		domErr := types.NewError(types.ErrBadRequest, "invalid request payload", err)
		apiErr, status := types.NewAPIError(domErr)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	req.ID = id
	if err := h.ucs.UpdateWorkorderByID(c, req.ToDomain()); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteWorkorderByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workorder ID"})
		return
	}

	if err := h.ucs.DeleteWorkorderByID(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListWorkorders(c *gin.Context) {
	filt := parseFilters(c)
	input := types.NewInput(c.Request)

	// Devuelve ([]domain.WorkorderListElement, types.PageInfo, error)
	list, pageInfo, err := h.ucs.ListWorkorders(c.Request.Context(), filt, input)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	// Usamos el helper del DTO para mapear y construir la respuesta
	resp := dto.FromDomainList(pageInfo, list)

	c.JSON(http.StatusOK, resp)
}

// ParseFilters extrae project_id, field_id, customer_id y campaign_id
func parseFilters(c *gin.Context) domain.WorkorderFilter {
	var f domain.WorkorderFilter
	if v := c.Query("project_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.ProjectID = &id
		}
	}
	if v := c.Query("field_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.FieldID = &id
		}
	}
	if v := c.Query("customer_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.CustomerID = &id
		}
	}
	if v := c.Query("campaign_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.CampaignID = &id
		}
	}
	return f
}

func (h *Handler) GetMetrics(c *gin.Context) {
	var filt domain.WorkorderFilter
	if v := c.Query("project_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project_id"})
			return
		}
		filt.ProjectID = &id
	}
	if v := c.Query("field_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field_id"})
			return
		}
		filt.FieldID = &id
	}
	if v := c.Query("customer_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer_id"})
			return
		}
		filt.CustomerID = &id
	}
	if v := c.Query("campaign_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid campaign_id"})
			return
		}
		filt.CampaignID = &id
	}
	m, err := h.ucs.GetMetrics(c.Request.Context(), filt)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportWorkorders(c *gin.Context) {
	filt := parseFilters(c)
	// Para exportación, usar un page_size muy grande para obtener todos los registros
	input := types.Input{
		Page:     1,
		PageSize: 100000, // Límite suficientemente grande para exportar todos
	}

	data, err := h.ucs.ExportWorkorders(c, filt, input)
	if err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}

	filename := workorderexcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
