package field

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/handler/dto"
)

// Handler encapsulates all dependencies for the Field HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Field handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registers all field routes.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/fields"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateField)       // Create a field
		public.GET("", h.ListFields)         // List all fields
		public.GET("/:id", h.GetField)       // Get a field by ID
		public.PUT("/:id", h.UpdateField)    // Update a field
		public.DELETE("/:id", h.DeleteField) // Delete a field
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

// CreateField handles POST /fields
func (h *Handler) CreateField(c *gin.Context) {
	var req dto.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	id, err := h.ucs.CreateField(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateFieldResponse{Message: "Field created", ID: id})
}

// ListFields handles GET /fields
func (h *Handler) ListFields(c *gin.Context) {
	fields, err := h.ucs.ListFields(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	dtos := make([]dto.Field, len(fields))
	for i, f := range fields {
		dtos[i] = dto.FromDomain(f)
	}
	c.JSON(http.StatusOK, dtos)
}

// GetField handles GET /fields/:id
func (h *Handler) GetField(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	f, err := h.ucs.GetField(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(*f))
}

// UpdateField handles PUT /fields/:id
func (h *Handler) UpdateField(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.UpdateField
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateField(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Field updated"})
}

// DeleteField handles DELETE /fields/:id
func (h *Handler) DeleteField(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.ucs.DeleteField(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Field deleted"})
}
