package labor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	labexcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/excel"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type UseCasesPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error)
	DeleteLabor(context.Context, int64) error
	UpdateLabor(context.Context, *domain.Labor) error
	ListLaborCategoriesByTypeID(context.Context, int64) ([]domain.LaborCategory, error)
	ListLaborByWorkorder(context.Context, int64) ([]domain.LaborRawItem, error)
	ListGroupLaborByWorkorder(context.Context, types.Input, int64, int64) ([]domain.LaborListItem, types.PageInfo, error)
	ExportGroupLaborXLSX(context.Context, types.Input, int64, int64) ([]byte, error)
	ExportAllGroupLabors(context.Context) ([]byte, error)
	GetMetrics(context.Context, domain.LaborFilter) (*domain.LaborMetrics, error)
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

// Handler encapsulates all dependencies for the LeaseType HTTP handler.
type Handler struct {
	ucs  UseCasesPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
	ucps project.UseCasesPort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort, up project.UseCasesPort) *Handler {
	return &Handler{
		ucs:  u,
		gsv:  s,
		acf:  c,
		mws:  m,
		ucps: up,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	public := r.Group(baseURL + "/projects/:id/labors")
	{
		public.POST("", h.CreateLabor)
		public.GET("", h.ListLabor)
		public.DELETE("/:idLabor", h.DeleteLabor)
		public.PUT("/:idLabor", h.UpdateLabor)
		public.GET("/labor-categories/:typeId", h.ListLaborCategories)
	}

	workorderGroup := r.Group(baseURL + "/labors")
	{
		workorderGroup.DELETE("/:idLabor", h.DeleteLaborByID)
		workorderGroup.GET("/:workorderID", h.ListLaborByWorkorder)
		workorderGroup.GET("/group/:projectID", h.ListGroupLaborByProject)
		workorderGroup.GET("/export/:projectID", h.ExportGroupLaborXLSX)
		workorderGroup.GET("/export/all", h.ExportAllGroupLabors)
		workorderGroup.GET("/metrics", h.GetMetrics)
	}
}

func (h *Handler) CreateLabor(c *gin.Context) {
	var req dto.LaborList
	projectIDStr := c.Param("id")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	_, err = h.ucps.GetProject(ctx, projectID)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	var labors []dto.CreateLabor

	for _, labor := range req.Labors {
		var laborResponse dto.CreateLabor

		laborID, err := h.ucs.CreateLabor(ctx, labor.ToDomain(projectID, userID))
		if err != nil {
			laborResponse = dto.CreateLabor{
				LaborID:     0,
				LaborName:   labor.Name,
				IsSaved:     false,
				ErrorDetail: err.Error(),
			}
		} else {
			laborResponse = dto.CreateLabor{
				LaborID:   laborID,
				LaborName: labor.Name,
				IsSaved:   true,
			}
		}

		labors = append(labors, laborResponse)
	}

	c.JSON(http.StatusMultiStatus, dto.CreateLaborsResponse{
		Message: "labors created",
		Labors:  labors,
	})
}

func (h *Handler) ListLabor(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))

	projectIDStr := c.Param("id")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	items, total, err := h.ucs.ListLabor(c.Request.Context(), page, perPage, projectID)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewListLaborsResponse(items, page, perPage, total)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateLabor(c *gin.Context) {
	projectIDStr := c.Param("id")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("idLabor"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor id"})
		return
	}
	var req dto.Labor
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateLabor(c.Request.Context(), req.ToDomain(projectID, userID)); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor updated successfully"})
}

func (h *Handler) DeleteLabor(c *gin.Context) {
	projectIDStr := c.Param("id")

	_, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("idLabor"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor id"})
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor deleted successfully"})
}

func (h *Handler) DeleteLaborByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("idLabor"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor id"})
		return
	}
	if err := h.ucs.DeleteLabor(c.Request.Context(), id); err != nil {
		apiErr, status := types.NewAPIError(err)
		c.JSON(status, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Labor deleted successfully"})
}

func (h *Handler) ListLaborCategories(c *gin.Context) {
	projectIDStr := c.Param("id")

	_, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	id, err := strconv.ParseInt(c.Param("typeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid labor type id"})
		return
	}

	laborCategories, err := h.ucs.ListLaborCategoriesByTypeID(c, id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.NewLaborCategoriesListResponse(laborCategories)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListLaborByWorkorder(c *gin.Context) {
	workorderID, ok := parseParamID(c, "workorderID")
	if !ok {
		return
	}

	items, err := h.ucs.ListLaborByWorkorder(c.Request.Context(), workorderID)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.ToLaborListResponse(items)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListGroupLaborByProject(c *gin.Context) {
	projectID, ok := parseParamID(c, "projectID")
	if !ok {
		return
	}

	fieldIDParam := c.Query("fieldID")
	if fieldIDParam == "" && projectID == 0 {
		apiErr, _ := types.NewAPIError(fmt.Errorf("fieldID or projectID is required"))
		c.Error(apiErr).SetMeta(map[string]any{"details": "fieldID or projectID requires a value"})
		return
	}

	var fieldID int64
	if fieldIDParam != "" {
		var err error
		fieldID, err = strconv.ParseInt(fieldIDParam, 10, 64)
		if err != nil {
			apiErr, _ := types.NewAPIError(fmt.Errorf("fieldID is not a valid integer"))
			c.Error(apiErr).SetMeta(map[string]any{"details": "fieldID is not a valid integer"})
			return
		}
	}

	input := types.NewInput(c.Request)

	list, pageInfo, err := h.ucs.ListGroupLaborByWorkorder(c.Request.Context(), input, projectID, fieldID)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}

	resp := dto.FromDomainList(pageInfo, list)

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ExportGroupLaborXLSX(c *gin.Context) {
	projectID, ok := parseParamID(c, "projectID")
	if !ok {
		return
	}

	fieldIDParam := c.Query("fieldID")
	if fieldIDParam == "" && projectID == 0 {
		types.NewErrorResponseHelper().BadRequest(c, "fieldID or projectID requires a value", nil)
		return
	}

	var fieldID int64
	if fieldIDParam != "" {
		var err error
		fieldID, err = strconv.ParseInt(fieldIDParam, 10, 64)
		if err != nil {
			types.NewErrorResponseHelper().BadRequest(c, "fieldID is not a valid integer", nil)
			return
		}
	}

	// Para exportación, usar un page_size muy grande para obtener todos los registros
	input := types.Input{
		Page:     1,
		PageSize: 100000, // Límite suficientemente grande para exportar todos
	}

	data, err := h.ucs.ExportGroupLaborXLSX(c.Request.Context(), input, projectID, fieldID)
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := labexcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) GetMetrics(c *gin.Context) {
	var filt domain.LaborFilter
	if v := c.Query("project_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project_id"})
			return
		}
		filt.ProjectID = &id
	}
	// Permitir field_id=0 o "null" para indicar "todos los campos"
	if v := c.Query("field_id"); v != "" && v != "null" && v != "undefined" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field_id"})
			return
		}
		// Solo setear field_id si es > 0 (field_id=0 = todos los campos)
		if id > 0 {
			filt.FieldID = &id
		}
	}
	m, err := h.ucs.GetMetrics(c.Request.Context(), filt)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomainMetrics(m))
}

func (h *Handler) ExportAllGroupLabors(c *gin.Context) {
	data, err := h.ucs.ExportAllGroupLabors(c.Request.Context())
	if err != nil {
		types.NewErrorResponseHelper().HandleDomainError(c, err)
		return
	}

	filename := "tabla_labores.xlsx"

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ----- HELPER -----

func parseParamID(c *gin.Context, param string) (int64, bool) {
	raw := strings.TrimSpace(c.Param(param))
	if raw == "" {
		apiErr := types.NewError(types.ErrInvalidID, param+" is required", nil)
		c.Error(apiErr).SetMeta(map[string]any{"param": param})
		return 0, false
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		apiErr := types.NewError(types.ErrInvalidID, param+" must be a positive integer", err)
		c.Error(apiErr).SetMeta(map[string]any{"param": param, "value": raw})
		return 0, false
	}

	return id, true
}
