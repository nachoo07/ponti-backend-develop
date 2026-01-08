package stock

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	stockExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/excel"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type UseCasesPort interface {
	GetStocksSummary(context.Context, int64, time.Time) ([]*domain.Stock, error)
	GetStocksPeriods(context.Context, int64) ([]string, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProject(context.Context, int64, int64, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockById(context.Context, int64) (*domain.Stock, error)
	GetLastStockByProjectId(context.Context, int64, int64) (*domain.Stock, bool, error)
	ExportAllStocks(ctx context.Context) ([]byte, error)
	UpdateUnitsConsumed(context.Context, domain.Stock, decimal.Decimal) error
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
type Handler struct {
	ucs  UseCasesPort
	gsv  GinEnginePort
	acf  ConfigAPIPort
	mws  MiddlewaresEnginePort
	ucps project.UseCasesPort
}

func NewHandler(
	u UseCasesPort,
	s GinEnginePort,
	c ConfigAPIPort,
	m MiddlewaresEnginePort,
	ucps project.UseCasesPort,
) *Handler {
	return &Handler{
		ucs:  u,
		gsv:  s,
		acf:  c,
		mws:  m,
		ucps: ucps,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL()

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	publicExcel := r.Group(baseURL + "/stocks/export")
	{
		publicExcel.GET("/all", h.ExportAllStocks)
	}

	public := r.Group(baseURL + "/projects/:id/stocks")
	{
		public.GET("/summary", h.getStocksSummary)
		public.GET("/periods", h.getStocksPeriods)
		public.PUT("/close-date", h.UpdateStocksCloseDate)
		public.PUT("/real-stock/:stockId", h.UpdateRealStock)
	}
}

func (h *Handler) getStocksSummary(c *gin.Context) {
	ctx := c.Request.Context()
	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	_, err = h.ucps.GetProject(ctx, projectId)
	if handleError(err, c) {
		return
	}

	cutoffDateStr := c.Query("cutoff_date")
	var cutoffDate time.Time
	if cutoffDateStr != "" {
		cutoffDate, err = time.Parse("2006-01-02", cutoffDateStr)
		if handleError(err, c) {
			return
		}
	}

	stocks, err := h.ucs.GetStocksSummary(ctx, projectId, cutoffDate)
	if handleError(err, c) {
		return
	}

	resp := dto.NewGetStocksListed(stocks)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getStocksPeriods(c *gin.Context) {
	ctx := c.Request.Context()
	projectIdStr := c.Param("id")

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	periods, err := h.ucs.GetStocksPeriods(ctx, projectId)
	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, periods)
}

// UpdateStocksCloseDate actualiza el close_date de los stocks por proyecto y field
func (h *Handler) UpdateStocksCloseDate(c *gin.Context) {
	ctx := c.Request.Context()
	projectIdStr := c.Param("id")

	monthPeriod, err := getMonthPeriod(c)
	if handleError(err, c) {
		return
	}

	yearPeriod, err := getYearPeriod(c)
	if handleError(err, c) {
		return
	}

	var req dto.UpdateCloseDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	userID, err := sharedmodels.ConvertStringToID(c)
	if handleError(err, c) {
		return
	}

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if handleError(err, c) {
		return
	}

	_, err = h.ucps.GetProject(ctx, projectId)
	if handleError(err, c) {
		return
	}

	if err := req.Validate(); handleError(err, c) {
		return
	}

	err = h.ucs.UpdateCloseDateByProject(
		ctx,
		projectId,
		monthPeriod,
		yearPeriod,
		req.ToDomain(&userID),
	)

	if handleError(err, c) {
		return
	}

	c.JSON(http.StatusOK, dto.UpdateCloseDateResponse{Message: "close_date updated successfully"})
}

func (h *Handler) UpdateRealStock(c *gin.Context) {
	ctx := c.Request.Context()
	stockIdStr := c.Param("stockId")

	var req dto.UpdateRealStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := sharedmodels.ConvertStringToID(c)
	if handleError(err, c) {
		return
	}

	stockId, err := strconv.ParseInt(stockIdStr, 10, 64)
	if handleError(err, c) {
		return
	}
	stockDomain, err := h.ucs.GetStockById(ctx, stockId)
	if handleError(err, c) {
		return
	}

	stockDomain.RealStockUnits = req.RealStockUnits
	stockDomain.UpdatedBy = &userID

	err = h.ucs.UpdateRealStockUnits(ctx, stockId, stockDomain)
	if handleError(err, c) {
		return
	}
	c.JSON(http.StatusOK, dto.NewUpdateRealStockResponse("real stock updated successfully"))
}

func handleError(err error, c *gin.Context) bool {
	if err == nil {
		return false
	}
	apiErr, _ := types.NewAPIError(err)
	c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
	return true
}

func getMonthPeriodOrDefault(c *gin.Context) (int64, error) {
	monthPeriodStr := c.Query("month_period")
	if monthPeriodStr == "" {
		return int64(time.Now().Month()), nil
	}
	monthPeriod, err := strconv.ParseInt(monthPeriodStr, 10, 64)
	if err != nil {
		return 0, types.NewError(types.ErrValidation, "Month period is in invalid format", nil)
	}

	if monthPeriod < 1 || monthPeriod > 12 {
		return 0, types.NewError(types.ErrValidation, "Month period must be between 1 and 12", nil)
	}
	return monthPeriod, nil
}

func getYearPeriodOrDefault(c *gin.Context) (int64, error) {
	yearPeriodStr := c.Query("year_period")
	if yearPeriodStr == "" {
		return int64(time.Now().Year()), nil
	}
	yearPeriod, err := strconv.ParseInt(yearPeriodStr, 10, 64)
	if err != nil {
		return 0, err
	}

	if yearPeriod < 0 {
		return 0, types.NewError(types.ErrValidation, "Year period must be greater than 0", nil)
	}

	return yearPeriod, nil
}

func getMonthPeriod(c *gin.Context) (int64, error) {
	monthPeriodStr := c.Query("month_period")

	if monthPeriodStr == "" {
		return 0, types.NewMissingFieldError("month_period")
	}
	return getMonthPeriodOrDefault(c)
}

func getYearPeriod(c *gin.Context) (int64, error) {
	yearPeriodStr := c.Query("year_period")

	if yearPeriodStr == "" {
		return 0, types.NewMissingFieldError("year_period")
	}
	return getYearPeriodOrDefault(c)
}

func (h *Handler) ExportAllStocks(c *gin.Context) {
	data, err := h.ucs.ExportAllStocks(c)
	if handleError(err, c) {
		return
	}

	filename := stockExcel.DefaultFilename

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
