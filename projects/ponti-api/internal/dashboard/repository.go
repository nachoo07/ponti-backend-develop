package dashboard

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db     GormEnginePort
	mapper *models.DashboardModelMapper
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db:     db,
		mapper: models.NewDashboardModelMapper(),
	}
}

func (r *Repository) GetDashboard(ctx context.Context, filter domain.DashboardFilter) (*domain.DashboardData, error) {
	// Determinar projectIDs una sola vez
	projectIDs, err := r.resolveProjectIDs(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Si no hay proyectos, retornar datos vacíos
	if len(projectIDs) == 0 {
		return r.createEmptyDashboardData(), nil
	}

	// Obtener todas las métricas principales en una sola query
	metricsData, err := r.getMetrics(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	// Obtener datos 1:N en paralelo (contribuciones, cultivos)
	contributionsData, err := r.getContributionsProgress(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	managementBalanceData, err := r.getManagementBalance(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	cropIncidenceData, err := r.getCropIncidence(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	operationalIndicatorsData, err := r.getOperationalIndicators(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	// Construir modelo de datos consolidado
	tempData := &models.DashboardDataModel{
		SowingHectares:         metricsData.SowingHectares,
		SowingTotalHectares:    metricsData.SowingTotalHectares,
		SowingProgressPercent:  metricsData.SowingProgressPct,
		CostsExecutedUSD:       metricsData.ExecutedCostsUSD,
		CostsBudgetUSD:         metricsData.BudgetCostUSD,
		CostsProgressPct:       metricsData.CostsProgressPct,
		HarvestHectares:        metricsData.HarvestHectares,
		HarvestTotalHectares:   metricsData.HarvestTotalHectares,
		HarvestProgressPercent: metricsData.HarvestProgressPct,
		OperatingResultUSD:     metricsData.OperatingResultUSD,
		OperatingTotalCostsUSD: metricsData.OperatingResultTotalCostsUSD,
		OperatingResultPct:     metricsData.OperatingResultPct,
	}

	// Retornar datos mapeados a dominio (contributionsData ya contiene contributions_progress_pct)
	return r.mapper.DashboardDataToDomain(tempData, cropIncidenceData, contributionsData, managementBalanceData, operationalIndicatorsData), nil
}

// resolveProjectIDs determina los IDs de proyectos a consultar basándose en los filtros
func (r *Repository) resolveProjectIDs(ctx context.Context, filter domain.DashboardFilter) ([]int64, error) {
	// Si tenemos ProjectID directamente, usarlo
	if filter.ProjectID != nil {
		return []int64{*filter.ProjectID}, nil
	}

	// Buscar proyectos relacionados con los otros filtros
	return r.getRelatedProjectIDs(ctx, filter)
}

// getRelatedProjectIDs encuentra los IDs de proyectos relacionados con los filtros
func (r *Repository) getRelatedProjectIDs(ctx context.Context, filter domain.DashboardFilter) ([]int64, error) {
	query := r.db.Client().WithContext(ctx).
		Table("projects p").
		Select("DISTINCT p.id").
		Where("p.deleted_at IS NULL")

	// Aplicar filtros dinámicamente
	if filter.CustomerID != nil {
		query = query.Where("p.customer_id = ?", *filter.CustomerID)
	}
	if filter.CampaignID != nil {
		query = query.Where("p.campaign_id = ?", *filter.CampaignID)
	}
	if filter.FieldID != nil {
		query = query.Where("EXISTS (SELECT 1 FROM fields f WHERE f.id = ? AND f.project_id = p.id AND f.deleted_at IS NULL)", *filter.FieldID)
	}

	var projectIDs []int64
	if err := query.Pluck("p.id", &projectIDs).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get related project IDs", err)
	}

	return projectIDs, nil
}

// createEmptyDashboardData retorna una estructura de dashboard vacía con valores por defecto
func (r *Repository) createEmptyDashboardData() *domain.DashboardData {
	zero := decimal.Zero
	tempData := &models.DashboardDataModel{
		SowingHectares:         zero,
		SowingTotalHectares:    zero,
		SowingProgressPercent:  zero,
		CostsExecutedUSD:       zero,
		CostsBudgetUSD:         zero,
		CostsProgressPct:       zero,
		HarvestHectares:        zero,
		HarvestTotalHectares:   zero,
		HarvestProgressPercent: zero,
		OperatingResultUSD:     zero,
		OperatingTotalCostsUSD: zero,
		OperatingResultPct:     zero,
	}

	return r.mapper.DashboardDataToDomain(
		tempData,
		[]models.CropIncidenceModel{},
		[]models.ContributionsProgressModel{},
		&models.ManagementBalanceModel{
			Summary: &models.ManagementBalanceSummary{
				IncomeUSD:                 zero,
				DirectCostsExecutedUSD:    zero,
				DirectCostsInvestedUSD:    zero,
				StockUSD:                  zero,
				RentExecutedUSD:           zero,
				RentUSD:                   zero,
				StructureExecutedUSD:      zero,
				StructureUSD:              zero,
				OperatingResultUSD:        zero,
				OperatingResultPct:        zero,
				SemillaCostUSD:            zero,
				InsumosCostUSD:            zero,
				LaboresCostUSD:            zero,
				SemillasInvertidosUSD:     zero,
				SemillasStockUSD:          zero,
				AgroquimicosInvertidosUSD: zero,
				AgroquimicosStockUSD:      zero,
				LaboresInvertidosUSD:      zero,
			},
			Breakdown: []models.ManagementBalanceBreakdown{},
			TotalsRow: &models.ManagementBalanceTotals{
				TotalExecutedUSD: zero,
				TotalInvestedUSD: zero,
				TotalStockUSD:    zero,
			},
		},
		&models.OperationalIndicatorModel{},
	)
}

// getMetrics obtiene todas las métricas principales del dashboard en una sola query
func (r *Repository) getMetrics(ctx context.Context, projectIDs []int64) (*models.DashboardMetricsModel, error) {
	var result models.DashboardMetricsModel

	err := r.db.Client().WithContext(ctx).
		Table("v3_dashboard_metrics").
		Select("*").
		Where("project_id IN ?", projectIDs).
		Scan(&result).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard metrics", err)
	}

	return &result, nil
}

// getContributionsProgress obtiene los datos del avance de aportes por inversor
func (r *Repository) getContributionsProgress(ctx context.Context, projectIDs []int64) ([]models.ContributionsProgressModel, error) {
	var results []models.ContributionsProgressModel

	err := r.db.Client().WithContext(ctx).
		Table("v3_dashboard_contributions_progress").
		Select("*").
		Where("project_id IN ?", projectIDs).
		Order("investor_id").
		Scan(&results).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get contributions progress data", err)
	}

	return results, nil
}

// getManagementBalance obtiene los datos del balance de gestión
func (r *Repository) getManagementBalance(ctx context.Context, projectIDs []int64) (*models.ManagementBalanceModel, error) {
	var summary models.ManagementBalanceSummary

	err := r.db.Client().WithContext(ctx).
		Table("v3_dashboard_management_balance").
		Select("*").
		Where("project_id IN ?", projectIDs).
		Scan(&summary).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get management balance data", err)
	}

	return &models.ManagementBalanceModel{
		Summary:   &summary,
		Breakdown: []models.ManagementBalanceBreakdown{}, // TODO: Implementar cuando se requiera
		TotalsRow: &models.ManagementBalanceTotals{
			TotalExecutedUSD: summary.DirectCostsExecutedUSD,
			TotalInvestedUSD: summary.DirectCostsInvestedUSD,
			TotalStockUSD:    summary.StockUSD,
		},
	}, nil
}

// getCropIncidence obtiene los datos de incidencia de costos por cultivo
func (r *Repository) getCropIncidence(ctx context.Context, projectIDs []int64) ([]models.CropIncidenceModel, error) {
	var results []models.CropIncidenceModel

	err := r.db.Client().WithContext(ctx).
		Table("v3_dashboard_crop_incidence").
		Select("*").
		Where("project_id IN ?", projectIDs).
		Order("crop_name").
		Scan(&results).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get crop incidence data", err)
	}

	return results, nil
}

// getOperationalIndicators obtiene los indicadores operativos
func (r *Repository) getOperationalIndicators(ctx context.Context, projectIDs []int64) (*models.OperationalIndicatorModel, error) {
	var result models.OperationalIndicatorModel

	err := r.db.Client().WithContext(ctx).
		Table("v3_dashboard_operational_indicators").
		Select("*").
		Where("project_id IN ?", projectIDs).
		Limit(1).
		Scan(&result).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get operational indicators data", err)
	}

	return &result, nil
}
