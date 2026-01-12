package lot

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/handler/dto"
)

// Handler encapsulates dependencies for the Lot HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Lot handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{ucs: u, gsv: s, mws: m}
}

// Routes registers HTTP routes for lots.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/lots"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateLot)
		public.GET("", h.ListLots)
		public.GET("/:id", h.GetLot)
		public.PUT("/:id", h.UpdateLot)
		public.DELETE("/:id", h.DeleteLot)
	}

	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing)
	}
}

// ProtectedPing is a test endpoint for protected routes.
func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Protected Pong!"})
}

// CreateLot handles POST /lots
func (h *Handler) CreateLot(c *gin.Context) {
	var req dto.CreateLot
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	dom := req.Lot.ToDomain()
	newID, err := h.ucs.CreateLot(c.Request.Context(), dom)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateLotResponse{Message: "Lot created successfully", ID: newID})
}

// ListLots handles GET /lots
func (h *Handler) ListLots(c *gin.Context) {
	lots, err := h.ucs.ListLots(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lots)
}

// GetLot handles GET /lots/:id
func (h *Handler) GetLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}

	lot, err := h.ucs.GetLot(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lot)
}

// UpdateLot handles PUT /lots/:id
func (h *Handler) UpdateLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	var req dto.UpdateLot
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	dom := req.Lot.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateLot(c.Request.Context(), dom); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot updated successfully"})
}

// DeleteLot handles DELETE /lots/:id
func (h *Handler) DeleteLot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid lot id"})
		return
	}
	if err := h.ucs.DeleteLot(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Lot deleted successfully"})
}
