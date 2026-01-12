package supply

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	supplyExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/excel"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/handler/dto"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type UseCasesPort interface {
	CreateSupply(ctx context.Context, s *domain.Supply) (int64, error)
	CreateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error
	GetSupply(ctx context.Context, id int64) (*domain.Supply, error)
	UpdateSupply(ctx context.Context, s *domain.Supply) error
	DeleteSupply(ctx context.Context, id int64) error
	ListSuppliesPaginated(
		ctx context.Context,
		projectID int64,
		campaignID int64,
		page int,
		perPage int,
		mode string,
	) ([]domain.Supply, int64, error)
	UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error
	ExportTableSupplies(ctx context.Context, projectID int64) ([]byte, error)
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

// Handler encapsulates all dependencies for the Project HTTP handler.
type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewHandler creates a new Project handler.
func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// Routes registers all project routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/supplies"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL)
	{
		public.POST("", h.CreateSupply)
		public.POST("/bulk", h.CreateSuppliesBulk)
		public.GET("", h.ListSupplies)
		public.GET("/:id", h.GetSupply)
		public.GET("/export/all", h.ExportTableSupplies)
		public.PUT("/:id", h.UpdateSupply)
		public.DELETE("/:id", h.DeleteSupply)
		public.PUT("/bulk", h.UpdateSuppliesBulk)
	}
}

func (h *Handler) CreateSupply(c *gin.Context) {
	var req dto.Supply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	newID, err := h.ucs.CreateSupply(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Supply created successfully", "id": newID})
}

func (h *Handler) CreateSuppliesBulk(c *gin.Context) {
	var req []dto.Supply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.CreateSuppliesBulk(c, supplies); err != nil {
		code := http.StatusInternalServerError
		if types.IsErrInvalidInput(err) {
			code = http.StatusBadRequest
		} else if types.IsConflict(err) {
			code = http.StatusConflict
		}
		c.JSON(code, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Supplies created successfully"})
}

func (h *Handler) ListSupplies(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	mode := c.Query("mode") // "and" o "or"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "1000"))

	// Debes tener tu UC adaptado a paginar y devolver total
	supplies, total, err := h.ucs.ListSuppliesPaginated(c.Request.Context(), projectID, campaignID, page, perPage, mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	resp := dto.NewListSuppliesResponse(supplies, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	supply, err := h.ucs.GetSupply(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(supply))
}

func (h *Handler) UpdateSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	var req dto.Supply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateSupply(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supply updated successfully"})
}

func (h *Handler) DeleteSupply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supply id"})
		return
	}
	if err := h.ucs.DeleteSupply(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supply deleted successfully"})
}

func (h *Handler) UpdateSuppliesBulk(c *gin.Context) {
	var req []dto.Supply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	supplies := make([]domain.Supply, len(req))
	for i := range req {
		supplies[i] = *req[i].ToDomain()
	}
	if err := h.ucs.UpdateSuppliesBulk(c.Request.Context(), supplies); err != nil {
		code := http.StatusInternalServerError
		if types.IsErrInvalidInput(err) {
			code = http.StatusBadRequest
		}
		c.JSON(code, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Supplies updated successfully"})
}

func (h *Handler) ExportTableSupplies(c *gin.Context) {
	projectID, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)

	data, err := h.ucs.ExportTableSupplies(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	filename := supplyExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
