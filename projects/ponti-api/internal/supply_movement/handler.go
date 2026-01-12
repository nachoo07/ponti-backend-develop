package supply_movement

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	providerdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	supplyExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/excel"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/handler/dto"
	createsupplymovement "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/handler/dto/create_supply_movement"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
)

type UseCasesPort interface {
	GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error)
	UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error
	GetProviders(context.Context) ([]providerdomain.Provider, error)
	ExportSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]byte, error)
	DeleteSupplyMovement(context.Context, int64, int64) error
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

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	publicProviders := r.Group(baseURL + "/providers")
	{
		publicProviders.GET("", h.GetProviders)
	}

	public := r.Group(baseURL + "/projects/:id/supply-movements")
	{
		public.POST("", h.CreateSupplyMovement)
		public.DELETE("/:idSupplyMovement", h.DeleteSupplyMovement)
		public.GET("", h.GetSupplyMovementsByProjectID)
		public.GET("/export", h.ExportSupplyMovementsByProjectID)
	}
}

type Handler struct {
	ucs  UseCasesPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
	ucpp project.UseCasesPort
}

func NewHandler(
	u UseCasesPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
	ucpp project.UseCasesPort,
) *Handler {
	return &Handler{
		ucs:  u,
		gsv:  s,
		acf:  c,
		mws:  m,
		ucpp: ucpp,
	}
}

func (h *Handler) CreateSupplyMovement(c *gin.Context) {
	ctx := c.Request.Context()
	var req createsupplymovement.CreateSupplyMovementRequestBulk

	userID, err := sharedmodels.ConvertStringToID(c)
	if handleError(err, c) {
		return
	}

	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	var supplyMovementsResponse []createsupplymovement.CreateSupplyMovementResponse

	for _, supplyMovement := range req.SupplyMovements {
		var supplyMovementResponse createsupplymovement.CreateSupplyMovementResponse

		err = supplyMovement.Validate()
		if err != nil {
			supplyMovementResponse = createsupplymovement.NewErrorCreateSupplyMovementResponse(err.Error())
		} else {
			supplyMovementId, err := h.ucs.CreateSupplyMovement(ctx, supplyMovement.ToDomain(projectId, &userID))
			if err != nil {
				supplyMovementResponse = createsupplymovement.NewErrorCreateSupplyMovementResponse(err.Error())
			} else {
				supplyMovementResponse = createsupplymovement.NewSuccessfulCreateSupplyMovementResponse(supplyMovementId)
			}
		}

		supplyMovementsResponse = append(supplyMovementsResponse, supplyMovementResponse)
	}

	c.JSON(http.StatusMultiStatus, createsupplymovement.CreateSupplyMovementBulkResponse{
		SupplyMovements: supplyMovementsResponse,
	})
}

func (h *Handler) GetSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	supplyMovements, err := h.ucs.GetEntriesSupplyMovementsByProjectID(ctx, projectId)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, dto.NewGetEntrySupplyMovementsResponse(supplyMovements))
}

func (h *Handler) DeleteSupplyMovement(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if handleError(err, c) {
		return
	}

	supplyMovementStr := c.Param("idSupplyMovement")
	supplyMovementId, err := strconv.ParseInt(supplyMovementStr, 10, 64)
	if handleError(err, c) {
		return
	}

	err = h.ucs.DeleteSupplyMovement(ctx, id, supplyMovementId)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{Message: "supply movement deleted successfully"})

}

func (h *Handler) UpdateSupplyMovementById(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.UpdateSupplyMovementEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	supplyMovementStr := c.Param("idSupplyMovement")
	supplyMovementId, err := strconv.ParseInt(supplyMovementStr, 10, 64)
	if handleError(err, c) {
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if handleError(err, c) {
		return
	}

	supplyMovement, err := h.ucs.GetSupplyMovementByID(ctx, supplyMovementId)
	if handleError(err, c) {
		return
	}

	if err = req.Validate(); handleError(err, c) {
		return
	}
	err = h.ucs.UpdateSupplyMovement(ctx, req.ToDomain(projectId, &userID, supplyMovement))

	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, types.MessageResponse{Message: "supply movement updated successfully"})
}

func (h *Handler) GetProviders(c *gin.Context) {
	ctx := c.Request.Context()

	providers, err := h.ucs.GetProviders(ctx)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, dto.NewGetProvidersResponse(providers))
}

func handleError(err error, c *gin.Context) bool {
	if err == nil {
		return false
	}
	apiErr, _ := types.NewAPIError(err)
	c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
	return true
}

func (h *Handler) ExportSupplyMovementsByProjectID(c *gin.Context) {
	ctx := c.Request.Context()

	projectIdStr := c.Param("id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	data, err := h.ucs.ExportSupplyMovementsByProjectID(ctx, projectId)
	if handleError(err, c) {
		return
	}

	filename := supplyExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
