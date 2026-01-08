package invoice

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/gin-gonic/gin"
)

type UseCasePort interface {
	GetInvoiceByWorkOrder(context.Context, int64) (*domain.Invoice, error)
	CreateInvoice(context.Context, *domain.Invoice) (int64, error)
	UpdateInvoice(context.Context, *domain.Invoice) error
	DeleteInvoice(context.Context, int64) error
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
	baseURL := h.cfg.APIBaseURL() + "/invoice"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("/:workorderID", h.GetInvoiceByWorkOrder)
		public.POST("/:workorderID", h.CreateInvoice)
		public.PUT("/:workorderID", h.UpdateInvoice)
		public.DELETE("/:workorderID", h.DeleteInvoice)

	}
}

func (h *Handler) GetInvoiceByWorkOrder(c *gin.Context) {
	id, ok := parseParamID(c, "workorderID")
	if !ok {
		return
	}

	item, err := h.ucs.GetInvoiceByWorkOrder(c.Request.Context(), id)
	if err != nil {
		responseError(c, err)
		return
	}

	resp := dto.FromDomain(item)

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateInvoice(c *gin.Context) {
	id, ok := parseParamID(c, "workorderID")
	if !ok {
		return
	}

	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var body dto.InvoiceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrValidation, "invalid request body", err).Error(),
		})
		return
	}

	item := body.ToDomain(id, userID)

	id, err := h.ucs.CreateInvoice(c.Request.Context(), item)
	if err != nil {
		responseError(c, err)
		return
	}

	item.ID = id
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Invoice saved"})
}

func (h *Handler) UpdateInvoice(c *gin.Context) {
	id, ok := parseParamID(c, "workorderID")
	if !ok {
		return
	}

	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var body dto.InvoiceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrValidation, "invalid request body", err).Error(),
		})
		return
	}

	item := body.ToDomain(id, userID)
	item.ID = id

	if err := h.ucs.UpdateInvoice(c.Request.Context(), item); err != nil {
		responseError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{Message: "Invoice saved"})
}

func (h *Handler) DeleteInvoice(c *gin.Context) {
	id, ok := parseParamID(c, "workorderID")
	if !ok {
		return
	}

	if err := h.ucs.DeleteInvoice(c.Request.Context(), id); err != nil {
		responseError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ---HELPERS---

func parseParamID(c *gin.Context, param string) (int64, bool) {
	raw := c.Param(param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrInvalidID, param+" is required", err).Error(),
		})
		return 0, false
	}
	return id, true
}

func responseError(c *gin.Context, err error) {
	switch {
	case types.IsNotFound(err):
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
	case types.IsValidationError(err):
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
	case types.IsConflict(err):
		c.JSON(http.StatusConflict, types.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
	}
}

func parseUserID(c *gin.Context) (int64, bool) {
	UserID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: types.NewError(types.ErrAuthorization, "invalid userID", err).Error(),
		})
		return 0, false
	}
	return UserID, true
}
