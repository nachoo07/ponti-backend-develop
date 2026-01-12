package classtype

import (
	"context"
	"net/http"
	"strconv"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
	"github.com/gin-gonic/gin"
)

// UseCasesPort expects domain.ClassType, not dto.ClassType
type UseCasesPort interface {
	ListClassTypes(context.Context) ([]domain.ClassType, error)
	CreateClassType(context.Context, *domain.ClassType) (int64, error)
	UpdateClassType(context.Context, *domain.ClassType) error
	DeleteClassType(context.Context, int64) error
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

// Handler encapsulates all dependencies for the ClassType HTTP handler.
type Handler struct {
	classTypeUC UseCasesPort
	gsv         GinEnginePort
	acf         ConfigAPIPort
	mws         MiddlewaresEnginePort
}

// NewHandler creates a new ClassType handler.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		classTypeUC: u,
		gsv:         s,
		acf:         c,
		mws:         m,
	}
}

// Routes registers all class type routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/types"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	group := r.Group(baseURL)
	{
		group.GET("", h.ListClassTypes)
		group.POST("", h.CreateClassType)
		group.PUT("/:id", h.UpdateClassType)
		group.DELETE("/:id", h.DeleteClassType)
	}
}
func (h *Handler) ListClassTypes(c *gin.Context) {
	classTypes, err := h.classTypeUC.ListClassTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	out := make([]dto.ClassType, len(classTypes))
	for i := range classTypes {
		out[i] = *dto.FromDomain(&classTypes[i])
	}
	c.JSON(http.StatusOK, out)
}
func (h *Handler) CreateClassType(c *gin.Context) {
	var req dto.ClassType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	newID, err := h.classTypeUC.CreateClassType(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Class type created successfully", "id": newID})
}
func (h *Handler) UpdateClassType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid class type id"})
		return
	}
	var req dto.ClassType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.classTypeUC.UpdateClassType(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Class type updated successfully"})
}
func (h *Handler) DeleteClassType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid class type id"})
		return
	}
	if err := h.classTypeUC.DeleteClassType(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Class type deleted successfully"})
}
