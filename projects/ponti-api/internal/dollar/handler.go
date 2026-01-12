package dollar

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/gin-gonic/gin"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type UseCasePort interface {
	ListByProject(context.Context, int64) ([]domain.DollarAverage, error)
	CreateOrUpdateBulk(context.Context, []domain.DollarAverage) error
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
	ucs UseCasePort
	gsv GinEnginePort
	cfg ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasePort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		cfg: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.cfg.APIBaseURL() + "/projects/:id/dollar-values"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("", h.ListByProject)
		public.PUT("", h.CreateorUpdateBulk)
	}
}

func (h *Handler) ListByProject(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrInvalidID, "projectId is required", err).Error(),
		})
		return
	}

	// Llamo al caso de uso
	items, err := h.ucs.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		switch {
		case types.IsNotFound(err):
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		case types.IsValidationError(err):
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		}
		return
	}

	// mapeo los domains a DTOs
	resp := make([]dto.MonthResponse, len(items))
	for i, d := range items {
		resp[i] = dto.FromDomainMonth(&d)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateorUpdateBulk(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrInvalidID, "projectId is required", err).Error(),
		})
		return
	}

	// Parseo el body JSON a DTO
	var in dto.BulkDollarAverageRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrValidation, "invalid request body", err).Error(),
		})
		return
	}

	// Convierto el DTO a Slice de domain
	items := in.ToDomainSlice(projectID)
	if err := h.ucs.CreateOrUpdateBulk(c.Request.Context(), items); err != nil {
		switch {
		case types.IsValidationError(err):
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		case types.IsConflict(err):
			c.JSON(http.StatusConflict, types.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Dollar average saved"})
}
