package campaign

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type UseCasesPort interface {
	CreateCampaign(context.Context, *domain.Campaign) (int64, error)
	ListCampaigns(context.Context, int64, string) ([]domain.Campaign, error)
	GetCampaign(context.Context, int64) (*domain.Campaign, error)
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
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/campaigns"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.GET("", h.ListCampaigns)
	}
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	var customerID int64
	customerIDQuery := c.Query("customer_id")
	if customerIDQuery != "" {
		var err error
		customerID, err = strconv.ParseInt(customerIDQuery, 10, 64)
		if err != nil {
			apiErr, _ := types.NewAPIError(err)
			c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
			return
		}
	}

	campaigns, err := h.ucs.ListCampaigns(c.Request.Context(), customerID, c.Query("project_name"))
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaigns)
}
