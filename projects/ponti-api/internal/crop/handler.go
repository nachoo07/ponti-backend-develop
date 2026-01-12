package crop

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/handler/dto"
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
	apiBase := "/api/" + apiVersion + "/crops"
	publicPrefix := apiBase + "/public"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateCrop)
		public.GET("", h.ListCrops)
		public.GET("/:id", h.GetCrop)
		public.PUT("/:id", h.UpdateCrop)
		public.DELETE("/:id", h.DeleteCrop)
	}
}

func (h *Handler) CreateCrop(c *gin.Context) {
	var req dto.CreateCrop
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newID, err := h.ucs.CreateCrop(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateCropResponse{
		Message: "Crop created successfully",
		ID:      newID,
	})
}

// ListCrops retrieves all crops.
func (h *Handler) ListCrops(c *gin.Context) {
	crops, err := h.ucs.ListCrops(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, crops)
}

// GetCrop retrieves a crop by its ID.
func (h *Handler) GetCrop(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid crop id"})
		return
	}

	crop, err := h.ucs.GetCrop(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, crop)
}

// UpdateCrop updates an existing crop.
func (h *Handler) UpdateCrop(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid crop id"})
		return
	}
	var req dto.Crop
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateCrop(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Crop updated successfully"})
}

// DeleteCrop deletes a crop by its ID.
func (h *Handler) DeleteCrop(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid crop id"})
		return
	}
	if err := h.ucs.DeleteCrop(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Crop deleted successfully"})
}
