package manager

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/handler/dto"
)

// Handler encapsula todas las dependencias para el HTTP handler de Manager.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler crea un nuevo handler de Manager.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registra todas las rutas de manager.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/managers"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateManager)       // Crear un manager
		public.GET("", h.ListManagers)         // Listar todos los customers
		public.GET("/:id", h.GetManager)       // Obtener un manager por ID
		public.PUT("/:id", h.UpdateManager)    // Actualizar un manager
		public.DELETE("/:id", h.DeleteManager) // Eliminar un manager
	}

	// Rutas protegidas.
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing) // Endpoint de prueba protegido
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreateManager(c *gin.Context) {
	var req dto.CreateManager
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateManager(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateManagerResponse{
		Message:   "Manager created successfully",
		ManagerID: newID,
	})
}

func (h *Handler) ListManagers(c *gin.Context) {
	customers, err := h.ucs.ListManagers(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

func (h *Handler) GetManager(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid manager id"})
		return
	}

	manager, err := h.ucs.GetManager(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manager)
}

// UpdateManager actualiza un manager existente.
func (h *Handler) UpdateManager(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid manager id"})
		return
	}
	var req dto.Manager
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateManager(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Manager updated successfully"})
}

// DeleteManager elimina un manager por su ID.
func (h *Handler) DeleteManager(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid manager id"})
		return
	}
	if err := h.ucs.DeleteManager(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Manager deleted successfully"})
}
