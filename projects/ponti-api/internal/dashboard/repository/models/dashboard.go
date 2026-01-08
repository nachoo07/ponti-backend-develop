package models

import (
	"time"

	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

// DashboardModel representa el modelo de base de datos para dashboard
type DashboardModel struct {
	ID int64

	sharedmodels.Base
}

// DashboardMetricsModel representa las métricas principales del dashboard desde v3_dashboard_metrics
type DashboardMetricsModel struct {
	CustomerID int64 `gorm:"column:customer_id"`
	ProjectID  int64 `gorm:"column:project_id"`
	CampaignID int64 `gorm:"column:campaign_id"`

	// Métricas de siembra
	SowingHectares      decimal.Decimal `gorm:"column:sowing_hectares"`
	SowingTotalHectares decimal.Decimal `gorm:"column:sowing_total_hectares"`
	SowingProgressPct   decimal.Decimal `gorm:"column:sowing_progress_pct"`

	// Métricas de cosecha
	HarvestHectares      decimal.Decimal `gorm:"column:harvest_hectares"`
	HarvestTotalHectares decimal.Decimal `gorm:"column:harvest_total_hectares"`
	HarvestProgressPct   decimal.Decimal `gorm:"column:harvest_progress_pct"`

	// Métricas de costos
	ExecutedCostsUSD decimal.Decimal `gorm:"column:executed_costs_usd"`
	BudgetCostUSD    decimal.Decimal `gorm:"column:budget_cost_usd"`
	CostsProgressPct decimal.Decimal `gorm:"column:costs_progress_pct"`

	// Resultado operativo
	OperatingResultIncomeUSD     decimal.Decimal `gorm:"column:operating_result_income_usd"`
	OperatingResultUSD           decimal.Decimal `gorm:"column:operating_result_usd"`
	OperatingResultTotalCostsUSD decimal.Decimal `gorm:"column:operating_result_total_costs_usd"`
	OperatingResultPct           decimal.Decimal `gorm:"column:operating_result_pct"`

	ProjectTotalHectares decimal.Decimal `gorm:"column:project_total_hectares"`
}

// DashboardDataModel representa el modelo de datos del dashboard consolidado
type DashboardDataModel struct {
	// Métricas de siembra
	SowingHectares        decimal.Decimal `db:"sowing_hectares"`
	SowingTotalHectares   decimal.Decimal `db:"sowing_total_hectares"`
	SowingProgressPercent decimal.Decimal `db:"sowing_progress_percent"`

	// Métricas de cosecha
	HarvestHectares        decimal.Decimal `db:"harvest_hectares"`
	HarvestTotalHectares   decimal.Decimal `db:"harvest_total_hectares"`
	HarvestProgressPercent decimal.Decimal `db:"harvest_progress_percent"`

	// Métricas de costos
	CostsExecutedUSD decimal.Decimal `db:"costs_executed_usd"`
	CostsBudgetUSD   decimal.Decimal `db:"costs_budget_usd"`
	CostsProgressPct decimal.Decimal `db:"costs_progress_pct"`
	ExecutedCostsUSD decimal.Decimal `db:"executed_costs_usd"`
	BudgetCostUSD    decimal.Decimal `db:"budget_cost_usd"`

	// Resultado operativo
	OperatingIncomeUSD     decimal.Decimal `db:"operating_income_usd"`
	OperatingTotalCostsUSD decimal.Decimal `db:"operating_total_costs_usd"`
	OperatingResultUSD     decimal.Decimal `db:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `db:"operating_result_pct"`

	// Balance de gestión - Semilla
	SemillaEjecutadosUSD decimal.Decimal `db:"semilla_ejecutados_usd"`
	SemillaInvertidosUSD decimal.Decimal `db:"semilla_invertidos_usd"`
	SemillaStockUSD      decimal.Decimal `db:"semilla_stock_usd"`

	// Balance de gestión - Insumos
	InsumosEjecutadosUSD decimal.Decimal `db:"insumos_ejecutados_usd"`
	InsumosInvertidosUSD decimal.Decimal `db:"insumos_invertidos_usd"`
	InsumosStockUSD      decimal.Decimal `db:"insumos_stock_usd"`

	// Balance de gestión - Labores
	LaboresEjecutadosUSD decimal.Decimal `db:"labores_ejecutados_usd"`
	LaboresInvertidosUSD decimal.Decimal `db:"labores_invertidos_usd"`
	LaboresStockUSD      decimal.Decimal `db:"labores_stock_usd"`

	// Balance de gestión - Arriendo
	ArriendoInvertidosUSD decimal.Decimal `db:"arriendo_invertidos_usd"`

	// Balance de gestión - Estructura
	EstructuraInvertidosUSD decimal.Decimal `db:"estructura_invertidos_usd"`

	// Totales del Balance de Gestión
	CostosDirectosEjecutados decimal.Decimal `db:"costos_directos_ejecutados_usd"`
	CostosDirectosInvertidos decimal.Decimal `db:"costos_directos_invertidos_usd"`
	CostosDirectosStock      decimal.Decimal `db:"costos_directos_stock_usd"`
}

// SowingProgressModel representa el modelo de avance de siembra
type SowingProgressModel struct {
	Hectares      *decimal.Decimal `db:"sowing_hectares"`
	TotalHectares *decimal.Decimal `db:"sowing_total_hectares"`
	ProgressPct   *decimal.Decimal `db:"sowing_progress_pct"`
}

// HarvestProgressModel representa el modelo de avance de cosecha
type HarvestProgressModel struct {
	Hectares      *decimal.Decimal `db:"harvest_hectares"`
	TotalHectares *decimal.Decimal `db:"harvest_total_hectares"`
	ProgressPct   *decimal.Decimal `db:"harvest_progress_pct"`
}

// CostsProgressModel representa el modelo de avance de costos
type CostsProgressModel struct {
	ExecutedCostsUSD *decimal.Decimal `db:"executed_costs_usd"`
	BudgetCostUSD    *decimal.Decimal `db:"budget_cost_usd"`
	ProgressPct      *decimal.Decimal `db:"costs_progress_pct"`
}

// ContributionsProgressModel representa el modelo de avance de aportes
type ContributionsProgressModel struct {
	InvestorID               *int64           `gorm:"column:investor_id"`
	InvestorName             *string          `gorm:"column:investor_name"`
	InvestorPercentage       *decimal.Decimal `gorm:"column:investor_percentage_pct"`
	ContributionsProgressPct *decimal.Decimal `gorm:"column:contributions_progress_pct"`
}

// OperatingResultModel representa el modelo de resultado operativo
type OperatingResultModel struct {
	IncomeUSD     *decimal.Decimal `db:"income_usd"`
	TotalCostsUSD *decimal.Decimal `db:"operating_result_total_costs_usd"`
	ResultUSD     *decimal.Decimal `db:"operating_result_usd"`
	ResultPct     *decimal.Decimal `db:"operating_result_pct"`
}

// ManagementBalanceModel representa el modelo de balance de gestión
type ManagementBalanceModel struct {
	Summary   *ManagementBalanceSummary
	Breakdown []ManagementBalanceBreakdown
	TotalsRow *ManagementBalanceTotals
}

// ManagementBalanceSummary representa el resumen del balance de gestión
type ManagementBalanceSummary struct {
	IncomeUSD                  decimal.Decimal `gorm:"column:income_usd"`
	DirectCostsExecutedUSD     decimal.Decimal `gorm:"column:costos_directos_ejecutados_usd"`
	DirectCostsInvestedUSD     decimal.Decimal `gorm:"column:costos_directos_invertidos_usd"`
	StockUSD                   decimal.Decimal `gorm:"column:costos_directos_stock_usd"`
	RentExecutedUSD            decimal.Decimal `gorm:"column:arriendo_ejecutados_usd"`
	RentUSD                    decimal.Decimal `gorm:"column:arriendo_invertidos_usd"`
	StructureExecutedUSD       decimal.Decimal `gorm:"column:estructura_ejecutados_usd"`
	StructureUSD               decimal.Decimal `gorm:"column:estructura_invertidos_usd"`
	OperatingResultUSD         decimal.Decimal `gorm:"column:operating_result_usd"`
	OperatingResultPct         decimal.Decimal `gorm:"column:operating_result_pct"`
	SemillaCostUSD             decimal.Decimal `gorm:"column:semilla_cost"`
	InsumosCostUSD             decimal.Decimal `gorm:"column:insumos_cost"`
	FertilizantesCostUSD       decimal.Decimal `gorm:"column:fertilizantes_cost"`
	LaboresCostUSD             decimal.Decimal `gorm:"column:labores_cost"`
	SemillasInvertidosUSD      decimal.Decimal `gorm:"column:semillas_invertidos_usd"`
	SemillasStockUSD           decimal.Decimal `gorm:"column:semillas_stock_usd"`
	AgroquimicosInvertidosUSD  decimal.Decimal `gorm:"column:agroquimicos_invertidos_usd"`
	AgroquimicosStockUSD       decimal.Decimal `gorm:"column:agroquimicos_stock_usd"`
	FertilizantesInvertidosUSD decimal.Decimal `gorm:"column:fertilizantes_invertidos_usd"`
	FertilizantesStockUSD      decimal.Decimal `gorm:"column:fertilizantes_stock_usd"`
	LaboresInvertidosUSD       decimal.Decimal `gorm:"column:labores_invertidos_usd"`
}

// ManagementBalanceBreakdown representa el desglose del balance por categoría
type ManagementBalanceBreakdown struct {
	Category    string          `db:"category"`
	ExecutedUSD decimal.Decimal `db:"executed_usd"`
	InvestedUSD decimal.Decimal `db:"invested_usd"`
	StockUSD    decimal.Decimal `db:"stock_usd"`
}

// ManagementBalanceTotals representa los totales del balance de gestión
type ManagementBalanceTotals struct {
	TotalExecutedUSD decimal.Decimal `db:"total_executed_usd"`
	TotalInvestedUSD decimal.Decimal `db:"total_invested_usd"`
	TotalStockUSD    decimal.Decimal `db:"total_stock_usd"`
}

// CropIncidenceModel representa el modelo de incidencia de cultivos
type CropIncidenceModel struct {
	CropID       int64           `gorm:"column:current_crop_id"`
	Name         string          `gorm:"column:crop_name"`
	Hectares     decimal.Decimal `gorm:"column:crop_hectares"`
	IncidencePct decimal.Decimal `gorm:"column:crop_incidence_pct"`
	CostPerHa    decimal.Decimal `gorm:"column:cost_per_ha_usd"`
}

// InvestorContributionModel representa el modelo de contribución de inversores
type InvestorContributionModel struct {
	InvestorID   int64           `db:"investor_id"`
	InvestorName string          `db:"investor_name"`
	Percentage   decimal.Decimal `db:"investor_percentage_pct"`
}

// OperationalIndicatorModel representa el modelo de indicadores operativos
type OperationalIndicatorModel struct {
	FirstWorkorderDate   *time.Time `gorm:"column:start_date"`
	FirstWorkorderNumber *string    `gorm:"column:first_workorder_id"`
	LastWorkorderDate    *time.Time `gorm:"column:end_date"`
	LastWorkorderNumber  *string    `gorm:"column:last_workorder_id"`
	LastStockCountDate   *time.Time `gorm:"column:last_stock_count_date"`
	CampaignClosingDate  *time.Time `gorm:"column:campaign_closing_date"`
}
