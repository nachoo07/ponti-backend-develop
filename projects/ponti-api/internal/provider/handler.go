package provider

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
)

type RepositoryPort interface {
	GetProviders(context.Context) ([]*domain.Provider, error)
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

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	publicGroup := r.Group(baseURL + "/providers")
	{
		publicGroup.GET("", h.GetProviders)
	}
}

type Handler struct {
	repo RepositoryPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
}

func NewHandler(
	repo RepositoryPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
) *Handler {
	return &Handler{
		repo: repo,
		gsv:  s,
		acf:  c,
		mws:  m,
	}
}

func (h *Handler) GetProviders(c *gin.Context) {
	ctx := c.Request.Context()

	providers, err := h.repo.GetProviders(ctx)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, dto.NewGetProvidersResponse(providers))
}

func handleError(err error, c *gin.Context) bool {
	if err == nil {
		return false
	}
	apiErr, _ := types.NewAPIError(err)
	c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
	return true
}
