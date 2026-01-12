package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/users"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateUser)
		public.GET("", h.ListUsers)
		public.GET("/:id", h.GetUser)
		public.PUT("/:id", h.UpdateUser)
		public.DELETE("/:id", h.DeleteUser)
		public.POST("/follow", h.FollowUser)
		//public.DELETE("/unfollow", h.UnfollowUser)
		public.GET("/:id/followees", h.GetFolloweeUsers)
		public.GET("/:id/followers", h.GetFollowerUsers)
	}

	// Rutas validadas
	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)
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

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateUser
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	newUserID, err := h.ucs.CreateUser(ctx, req.ToDomain())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateUserResponse{
		Message: "User created successfully",
		UserID:  newUserID,
	})
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.ucs.ListUsers(c.Request.Context())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")

	person, err := h.ucs.GetUser(c.Request.Context(), id)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, person)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var updatedUser dto.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	if err := h.ucs.UpdateUser(c.Request.Context(), updatedUser.ToDomain()); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "User updated successfully",
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	hardDelete := c.Query("hardDelete") == "true"
	if err := h.ucs.DeleteUser(c.Request.Context(), id, hardDelete); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "User deleted successfully",
	})
}

func (h *Handler) FollowUser(c *gin.Context) {
	var req dto.Follow
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	relationID, err := h.ucs.FollowUser(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.FollowUserResponse{
		Message:        "Follow relationship created successfully",
		FollowRelation: relationID,
	})
}

func (h *Handler) GetFolloweeUsers(c *gin.Context) {
	followerID := c.Param("id")
	ctx := c.Request.Context()

	followees, err := h.ucs.GetFolloweeUsers(ctx, followerID)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, dto.GetFolloweesResponse{
		Message:   "Followees retrieved successfully",
		Followees: followees,
	})
}

// GetFollowerUsers obtiene la lista de usuarios que siguen al usuario indicado (followers).
func (h *Handler) GetFollowerUsers(c *gin.Context) {
	// Se utiliza el parámetro "id" para identificar al usuario cuyo listado de followers se desea obtener.
	followeeID := c.Param("id")
	ctx := c.Request.Context()

	// Se llama al caso de uso que devuelve la lista de followers.
	followers, err := h.ucs.GetFollowerUsers(ctx, followeeID)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	// Se envía la respuesta utilizando un DTO específico para followers.
	c.JSON(http.StatusOK, dto.GetFollowersResponse{
		Message:   "Followers retrieved successfully",
		Followers: followers,
	})
}
