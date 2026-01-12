package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/notification/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/notification"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.POST("", h.SendEmail)
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
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) SendEmail(c *gin.Context) {
	// cuil, err := mdw.ExtractClaim(c, "sub", "")
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, sup.ErrorResponse{
	// 		Error: err.Error(),
	// 	})
	// 	return
	// }

	var req dto.EmailVerification
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	if err := h.ucs.SendEmail(ctx, req.ToDomain().Address, req.ToDomain().Subject, req.ToDomain().Body); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Email verification email sent"})
}
