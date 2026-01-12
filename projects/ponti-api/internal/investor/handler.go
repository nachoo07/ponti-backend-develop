package investor

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/handler/dto"
)

// Handler encapsulates all dependencies for the Investor HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Investor handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registers all investor routes.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/investors"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateInvestor)       // Create an investor
		public.GET("", h.ListInvestors)         // List all investors
		public.GET("/:id", h.GetInvestor)       // Get an investor by ID
		public.PUT("/:id", h.UpdateInvestor)    // Update an investor
		public.DELETE("/:id", h.DeleteInvestor) // Delete an investor
	}

	// Protected routes.
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing) // Protected test endpoint
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

// CreateInvestor handles the creation of a new investor.
func (h *Handler) CreateInvestor(c *gin.Context) {
	var req dto.CreateInvestor
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateInvestor(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateInvestorResponse{
		Message:    "Investor created successfully",
		InvestorID: newID,
	})
}

// ListInvestors retrieves all investors.
func (h *Handler) ListInvestors(c *gin.Context) {
	investors, err := h.ucs.ListInvestors(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, investors)
}

// GetInvestor retrieves an investor by its ID.
func (h *Handler) GetInvestor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid investor id"})
		return
	}

	investor, err := h.ucs.GetInvestor(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, investor)
}

// UpdateInvestor updates an existing investor.
func (h *Handler) UpdateInvestor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid investor id"})
		return
	}
	var req dto.Investor
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateInvestor(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Investor updated successfully"})
}

// DeleteInvestor deletes an investor by its ID.
func (h *Handler) DeleteInvestor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid investor id"})
		return
	}
	if err := h.ucs.DeleteInvestor(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Investor deleted successfully"})
}
