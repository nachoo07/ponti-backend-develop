package customer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/handler/dto"
)

// Handler encapsula todas las dependencias para el HTTP handler de Customer.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler crea un nuevo handler de Customer.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

// Routes registra todas las rutas de customer.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/customers"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateCustomer)       // Crear un customer
		public.GET("", h.ListCustomers)         // Listar todos los customers
		public.GET("/:id", h.GetCustomer)       // Obtener un customer por ID
		public.PUT("/:id", h.UpdateCustomer)    // Actualizar un customer
		public.DELETE("/:id", h.DeleteCustomer) // Eliminar un customer
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

// CreateCustomer maneja la creación de un nuevo customer.
func (h *Handler) CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomer
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateCustomer(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateCustomerResponse{
		Message:    "Customer created successfully",
		CustomerID: newID,
	})
}

// ListCustomers recupera todos los customers.
func (h *Handler) ListCustomers(c *gin.Context) {
	customers, err := h.ucs.ListCustomers(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

// GetCustomer recupera un customer por su ID.
func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer id"})
		return
	}

	customer, err := h.ucs.GetCustomer(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// UpdateCustomer actualiza un customer existente.
func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer id"})
		return
	}
	var req dto.Customer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateCustomer(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Customer updated successfully"})
}

// DeleteCustomer elimina un customer por su ID.
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer id"})
		return
	}
	if err := h.ucs.DeleteCustomer(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Customer deleted successfully"})
}
