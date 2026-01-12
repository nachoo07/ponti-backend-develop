package person

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/handler/dto"
)

type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/person"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreatePerson)
		public.GET("", h.ListPersons)
		public.GET("/:id", h.GetPerson)
		public.PUT("/:id", h.UpdatePerson)
		public.DELETE("/:id", h.DeletePerson)
	}

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)
		// Puedes añadir rutas aquí si es necesario
	}

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)

		protected.GET("/ping", h.ProtectedPing)

		protected.POST("", h.CreatePerson)
		protected.GET("", h.ListPersons)
		protected.GET("/:id", h.GetPerson)
		protected.PUT("/:id", h.UpdatePerson)
		protected.DELETE("/:id", h.DeletePerson)

	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) CreatePerson(c *gin.Context) {
	var req dto.CreatePerson
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	newPersonID, err := h.ucs.CreatePerson(ctx, req.ToDomain())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.CreatePersonResponse{
		Message:  "Person created successfully",
		PersonID: newPersonID,
	})
}

func (h *Handler) DeletePerson(c *gin.Context) {
	// Extraer el ID de la URL
	id := c.Param("id")

	// Extraer el parámetro hardDelete del query string (opcional)
	hardDelete := c.Query("hardDelete") == "true"

	// Llamar al caso de uso con la información
	err := h.ucs.DeletePerson(c.Request.Context(), id, hardDelete)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Responder con éxito
	if hardDelete {
		c.JSON(http.StatusOK, types.MessageResponse{
			Message: "Person permanently deleted successfully",
		})
	} else {
		c.JSON(http.StatusOK, types.MessageResponse{
			Message: "Person deleted successfully",
		})
	}
}

func (h *Handler) UpdatePerson(c *gin.Context) {
	// Validamos el JSON de la solicitud en un DTO de actualización
	var req dto.UpdatePerson
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	if err := h.ucs.UpdatePerson(ctx, id, req.ToDomain()); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Person successfully updated",
	})
}

func (h *Handler) GetPerson(c *gin.Context) {
	id := c.Param("id")

	person, err := h.ucs.GetPerson(c.Request.Context(), id)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, person)
}

func (h *Handler) ListPersons(c *gin.Context) {
	persons, err := h.ucs.ListPersons(c.Request.Context())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	c.JSON(http.StatusOK, persons)
}
