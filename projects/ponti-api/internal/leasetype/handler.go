package leasetype

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
)

type UseCasesPort interface {
	CreateLeaseType(context.Context, *domain.LeaseType) (int64, error)
	ListLeaseTypes(context.Context) ([]domain.LeaseType, error)
	GetLeaseType(context.Context, int64) (*domain.LeaseType, error)
	UpdateLeaseType(context.Context, *domain.LeaseType) error
	DeleteLeaseType(context.Context, int64) error
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

// Handler encapsulates all dependencies for the LeaseType HTTP handler.
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
	baseURL := h.acf.APIBaseURL() + "/lease-types"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateLeaseType)
		public.GET("", h.ListLeaseTypes)
		public.GET("/:id", h.GetLeaseType)
		public.PUT("/:id", h.UpdateLeaseType)
		public.DELETE("/:id", h.DeleteLeaseType)
	}
}

func (h *Handler) CreateLeaseType(c *gin.Context) {
	var req dto.LeaseType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := &domain.LeaseType{Name: req.Name}
	id, err := h.ucs.CreateLeaseType(c.Request.Context(), dom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateLeaseTypeResponse{Message: "Lease type created", ID: id})
}

func (h *Handler) ListLeaseTypes(c *gin.Context) {
	list, err := h.ucs.ListLeaseTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	resp := make([]dto.LeaseType, len(list))
	for i, lt := range list {
		resp[i] = *dto.FromDomain(lt)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetLeaseType(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	lt, err := h.ucs.GetLeaseType(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(*lt))
}

func (h *Handler) UpdateLeaseType(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.LeaseType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateLeaseType(c.Request.Context(), req.ToDomain()); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lease type updated"})
}

func (h *Handler) DeleteLeaseType(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.ucs.DeleteLeaseType(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lease type deleted"})
}
