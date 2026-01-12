package stock

import (
	"context"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

// NewRepository crea una nueva instancia del repositorio de Stock
func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

// GetStocks retorna stocks filtrando por nombre de proyecto, nombre de field y opcionalmente por fecha de corte
func (r *Repository) GetStocks(ctx context.Context, projectId int64, closeDate time.Time) ([]*domain.Stock, error) {
	db := r.db.Client().WithContext(ctx)
	var t time.Time

	query := db.Model(&models.Stock{}).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("SupplyMovements").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Where("projects.id = ?", projectId)

	if closeDate != t {
		query = query.Where("stocks.close_date = ?", closeDate)
	} else {
		// Si no se especifica fecha, obtener el stock activo (sin filtro de fecha)
		query = query.Where("stocks.close_date IS NULL")
	}

	var stockModels []models.Stock
	if err := query.Order("stocks.id DESC").Find(&stockModels).Error; err != nil {
		return nil, err
	}

	// OPTIMIZACIÓN: Calcular consumed en una sola consulta para evitar N+1 problem
	if len(stockModels) > 0 {
		var consumedResults []struct {
			SupplyID int64           `gorm:"column:supply_id"`
			Consumed decimal.Decimal `gorm:"column:consumed"`
		}

		// Obtener todos los supply_ids
		supplyIDs := make([]int64, len(stockModels))
		for i, stock := range stockModels {
			supplyIDs[i] = stock.SupplyID
		}

		// Calcular consumed para todos los supplies en una sola consulta
		err := db.Model(&workordermodels.WorkorderItem{}).
			Joins("JOIN workorders ON workorders.id = workorder_items.workorder_id").
			Where("workorders.project_id = ? AND workorder_items.supply_id IN (?)", projectId, supplyIDs).
			Select("workorder_items.supply_id, COALESCE(SUM(workorder_items.total_used), 0) as consumed").
			Group("workorder_items.supply_id").
			Scan(&consumedResults).Error
		if err != nil {
			return nil, err
		}

		// Crear mapa de consumed por supply_id
		consumedMap := make(map[int64]decimal.Decimal)
		for _, result := range consumedResults {
			consumedMap[result.SupplyID] = result.Consumed
		}

		// Asignar consumed a cada stock
		for i := range stockModels {
			if consumed, exists := consumedMap[stockModels[i].SupplyID]; exists {
				stockModels[i].Consumed = consumed
			} else {
				stockModels[i].Consumed = decimal.Zero
			}
		}
	}

	stocks := make([]*domain.Stock, 0, len(stockModels))
	for i := range stockModels {
		stocks = append(stocks, stockModels[i].ToDomain())
	}
	return stocks, nil
}

func (r *Repository) GetStocksPeriods(ctx context.Context, projectId int64) ([]string, error) {
	var rawPeriods []time.Time

	err := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("project_id = ? AND close_date IS NOT NULL", projectId).
		Distinct("close_date").
		Pluck("close_date", &rawPeriods).Error
	if err != nil {
		return nil, err
	}

	periods := make([]string, len(rawPeriods))
	for i, t := range rawPeriods {
		periods[i] = t.Format("2006-01-02")
	}

	return periods, nil
}

func (r *Repository) CreateStock(ctx context.Context, stock *domain.Stock) (int64, error) {
	if stock == nil {
		return 0, types.NewError(types.ErrBadRequest, "stock is nil", nil)
	}
	model := models.FromDomain(stock)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create stock", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateCloseDateByProject(ctx context.Context, projectId int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateCloseDateFromDomain(stock)
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("project_id = ?", projectId).
		Where("close_date IS NULL").
		Updates(stockUpdate)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "no stocks found to update", nil)
	}
	return nil
}

func (r *Repository) UpdateRealStockUnits(ctx context.Context, stockId int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateRealUnitsFromDomain(stock)
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockId).
		Updates(stockUpdate)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "no stock found to update", nil)
	}
	return nil
}

func (r *Repository) UpdateUnitsConsumed(ctx context.Context, stockDomain domain.Stock, quantity decimal.Decimal) error {
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockDomain.ID).
		Update("units_consumed", gorm.Expr("units_consumed + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "no stock found to update", nil)
	}
	return nil
}

func (r *Repository) GetStockById(ctx context.Context, stockId int64) (*domain.Stock, error) {
	var stockModel models.Stock
	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		First(&stockModel, stockId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.NewError(types.ErrNotFound, "stock not found", nil)
		}
		return nil, err
	}
	return stockModel.ToDomain(), nil
}

func (r *Repository) GetLastStockByProjectId(ctx context.Context, projectId int64, supplyId int64) (*domain.Stock, bool, error) {
	var stockModel models.Stock
	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Where("project_id = ?", projectId).
		Where("supply_id = ?", supplyId).
		Where("close_date is null").
		First(&stockModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, true, nil
		}

		return nil, false, err
	}

	return stockModel.ToDomain(), false, nil
}

func (r *Repository) GetStockByPeriodAndProjectId(ctx context.Context, projectId int64) (*domain.Stock, error) {
	var stockModel models.Stock

	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Where("project_id = ?", projectId).
		Where("close_date IS NULL").
		First(&stockModel).Error
	if err != nil {
		return nil, err
	}

	return stockModel.ToDomain(), nil
}

func (r *Repository) ListAllStocks(ctx context.Context) ([]*domain.Stock, error) {
	var stockModel []models.Stock

	db := r.db.Client().WithContext(ctx)

	query := db.
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("SupplyMovements")

	if err := query.Find(&stockModel).Error; err != nil {
		return nil, err
	}

	// Calcular consumed para todos los stocks agrupando por project_id + supply_id
	if len(stockModel) > 0 {
		var consumedResults []struct {
			ProjectID int64           `gorm:"column:project_id"`
			SupplyID  int64           `gorm:"column:supply_id"`
			Consumed  decimal.Decimal `gorm:"column:consumed"`
		}

		// Calcular consumed agrupado por project_id y supply_id
		err := db.Model(&workordermodels.WorkorderItem{}).
			Joins("JOIN workorders ON workorders.id = workorder_items.workorder_id").
			Select("workorders.project_id, workorder_items.supply_id, COALESCE(SUM(workorder_items.total_used), 0) as consumed").
			Group("workorders.project_id, workorder_items.supply_id").
			Scan(&consumedResults).Error
		if err != nil {
			return nil, err
		}

		// Crear mapa de consumed por project_id:supply_id
		type stockKey struct {
			ProjectID int64
			SupplyID  int64
		}
		consumedMap := make(map[stockKey]decimal.Decimal)
		for _, result := range consumedResults {
			key := stockKey{ProjectID: result.ProjectID, SupplyID: result.SupplyID}
			consumedMap[key] = result.Consumed
		}

		// Asignar consumed a cada stock
		for i := range stockModel {
			key := stockKey{ProjectID: stockModel[i].ProjectID, SupplyID: stockModel[i].SupplyID}
			if consumed, exists := consumedMap[key]; exists {
				stockModel[i].Consumed = consumed
			} else {
				stockModel[i].Consumed = decimal.Zero
			}
		}
	}

	out := make([]*domain.Stock, 0, len(stockModel))
	for i := range stockModel {
		out = append(out, stockModel[i].ToDomain())
	}
	return out, nil
}
