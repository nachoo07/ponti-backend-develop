package commercialization

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/gin-gonic/gin"

	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type UseCasePort interface {
	CreateOrUpdateBulk(context.Context, []domain.CropCommercialization) error
	ListByProject(context.Context, int64) ([]domain.CropCommercialization, error)
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
	baseURL := h.cfg.APIBaseURL() + "/projects/:id/commercializations"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("", h.ListByProject)
		public.POST("", h.CreateOrUpdateBulk)
	}
}

// Listar por proyecto
func (h *Handler) ListByProject(c *gin.Context) {
	projectID, ok := parseParamID(c, "id")
	if !ok {
		return
	}

	items, err := h.ucs.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]dto.CommercializationResponse, len(items))
	for i, d := range items {
		resp[i] = dto.FromDomain(&d)
	}

	c.JSON(http.StatusOK, resp)
}

// Crear proyecto
func (h *Handler) CreateOrUpdateBulk(c *gin.Context) {
	projectID, ok := parseParamID(c, "id")
	if !ok {
		return
	}
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var body dto.BulkCommercializationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: types.NewError(types.ErrValidation, "invalid request body", err).Error(),
		})
		return
	}

	items := body.ToDomainSlice(projectID, userID)
	if err := h.ucs.CreateOrUpdateBulk(c.Request.Context(), items); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Crop commercialization saved"})
}

// ---- HELPERS -----

// Lee y valida el parametro en la ruta de tipo int64
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

// extrae userID usando Sharemodels.ConvertStringToID
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

// unifica el switch de errores
func respondError(c *gin.Context, err error) {
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
