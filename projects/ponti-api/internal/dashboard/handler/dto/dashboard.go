package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

// ===== REQUEST DTOs =====

// DashboardFilterRequest representa el filtro de request para el dashboard
type DashboardFilterRequest struct {
	CustomerID *int64 `json:"customer_id" binding:"omitempty"`
	ProjectID  *int64 `json:"project_id" binding:"omitempty"`
	CampaignID *int64 `json:"campaign_id" binding:"omitempty"`
	FieldID    *int64 `json:"field_id" binding:"omitempty"`
}

// DashboardRequest representa un request de dashboard (para casos de creación/actualización)
type DashboardRequest struct {
	Metrics               MetricsRequest               `json:"metrics"`
	ManagementBalance     ManagementBalanceRequest     `json:"management_balance"`
	CropIncidence         CropIncidenceRequest         `json:"crop_incidence"`
	OperationalIndicators OperationalIndicatorsRequest `json:"operational_indicators"`
}

// MetricsRequest representa las métricas en el request
type MetricsRequest struct {
	Sowing                SowingMetricRequest          `json:"sowing"`
	Harvest               HarvestMetricRequest         `json:"harvest"`
	Costs                 CostsMetricRequest           `json:"costs"`
	InvestorContributions InvestorContributionsRequest `json:"investor_contributions"`
	OperatingResult       OperatingResultMetricRequest `json:"operating_result"`
}

// SowingMetricRequest representa la métrica de siembra en el request
type SowingMetricRequest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// HarvestMetricRequest representa la métrica de cosecha en el request
type HarvestMetricRequest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// CostsMetricRequest representa la métrica de costos en el request
type CostsMetricRequest struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	BudgetUSD   decimal.Decimal `json:"budget_usd"`
}

// InvestorContributionsRequest representa la métrica de contribuciones en el request
type InvestorContributionsRequest struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	Breakdown   any             `json:"breakdown"`
}

// OperatingResultMetricRequest representa la métrica de resultado operativo en el request
type OperatingResultMetricRequest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	IncomeUSD     decimal.Decimal `json:"income_usd"`
	TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
}

// ManagementBalanceRequest representa el balance de gestión en el request
type ManagementBalanceRequest struct {
	Summary   BalanceSummaryRequest     `json:"summary"`
	Breakdown []BalanceBreakdownRequest `json:"breakdown"`
	TotalsRow BalanceTotalsRequest      `json:"totals_row"`
}

// BalanceSummaryRequest representa el resumen del balance en el request
type BalanceSummaryRequest struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	StockUSD               decimal.Decimal `json:"stock_usd"`
	RentUSD                decimal.Decimal `json:"rent_usd"`
	StructureUSD           decimal.Decimal `json:"structure_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
}

// BalanceBreakdownRequest representa el desglose por categoría en el request
type BalanceBreakdownRequest struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
}

// BalanceTotalsRequest representa la fila de totales en el request
type BalanceTotalsRequest struct {
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	InvestedUSD decimal.Decimal `json:"invested_usd"`
	StockUSD    decimal.Decimal `json:"stock_usd"`
}

// CropIncidenceRequest representa la incidencia de cultivos en el request
type CropIncidenceRequest struct {
	Crops []CropRequest    `json:"crops"`
	Total CropTotalRequest `json:"total"`
}

// CropRequest representa un cultivo en el request
type CropRequest struct {
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	RotationPct  decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// CropTotalRequest representa los totales de cultivos en el request
type CropTotalRequest struct {
	Hectares          decimal.Decimal `json:"hectares"`
	RotationPct       decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
}

// OperationalIndicatorsRequest representa los indicadores operativos en el request
type OperationalIndicatorsRequest struct {
	Cards []OperationalCardRequest `json:"cards"`
}

// OperationalCardRequest representa una tarjeta de indicador operativo en el request
type OperationalCardRequest struct {
	Key           string  `json:"key"`
	Title         string  `json:"title"`
	Date          *string `json:"date"`
	WorkorderID   any     `json:"workorder_id"`
	WorkorderCode any     `json:"workorder_code"`
	AuditID       any     `json:"audit_id"`
	AuditCode     any     `json:"audit_code"`
	Status        any     `json:"status"`
}

// ===== RESPONSE DTOs =====

// BalanceCategory representa categorías canónicas para items del balance
type BalanceCategory string

const (
	BalanceCategoryDirectCosts BalanceCategory = "DIRECT_COSTS"
	BalanceCategorySeed        BalanceCategory = "SEED"
	BalanceCategorySupplies    BalanceCategory = "SUPPLIES"
	BalanceCategoryFertilizers BalanceCategory = "FERTILIZERS"
	BalanceCategoryLabors      BalanceCategory = "LABORS"
	BalanceCategoryLease       BalanceCategory = "LEASE"
	BalanceCategoryAdmin       BalanceCategory = "ADMIN"
)

// AllowedBalanceCategories ayuda en validación y (des)serialización
var AllowedBalanceCategories = map[BalanceCategory]struct{}{
	BalanceCategoryDirectCosts: {},
	BalanceCategorySeed:        {},
	BalanceCategorySupplies:    {},
	BalanceCategoryFertilizers: {},
	BalanceCategoryLabors:      {},
	BalanceCategoryLease:       {},
	BalanceCategoryAdmin:       {},
}

// Valid verifica si la categoría es válida
func (c BalanceCategory) Valid() bool {
	_, ok := AllowedBalanceCategories[c]
	return ok
}

// roundToNearestInteger redondea un decimal al entero más cercano
func roundToNearestInteger(d decimal.Decimal) decimal.Decimal {
	return d.Round(0)
}

// DashboardResponse representa la respuesta del dashboard con la estructura exacta del JSON
type DashboardResponse struct {
	SchemaVersion         string                `json:"schema_version"`
	Metrics               Metrics               `json:"metrics"`
	ManagementBalance     ManagementBalance     `json:"management_balance"`
	CropIncidence         CropIncidence         `json:"crop_incidence"`
	OperationalIndicators OperationalIndicators `json:"operational_indicators"`
}

// Metrics representa las 5 cards de métricas
type Metrics struct {
	Sowing                SowingMetric          `json:"sowing"`
	Harvest               HarvestMetric         `json:"harvest"`
	Costs                 CostsMetric           `json:"costs"`
	InvestorContributions InvestorContributions `json:"investor_contributions"`
	OperatingResult       OperatingResultMetric `json:"operating_result"`
}

// SowingMetric representa la métrica de siembra
type SowingMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// HarvestMetric representa la métrica de cosecha
type HarvestMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// CostsMetric representa la métrica de costos
type CostsMetric struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	BudgetUSD   decimal.Decimal `json:"budget_usd"`
}

// InvestorContributions representa la métrica de contribuciones de inversores
type InvestorContributions struct {
	Items []InvestorItem `json:"items"`
}

// InvestorItem representa un item de inversor
type InvestorItem struct {
	InvestorID               int64           `json:"investor_id"`
	InvestorName             string          `json:"investor_name"`
	SharePct                 decimal.Decimal `json:"share_pct"`                  // % teórico acordado (abajo)
	ContributionsProgressPct decimal.Decimal `json:"contributions_progress_pct"` // % real de aportes (arriba)
}

// OperatingResultMetric representa la métrica de resultado operativo
type OperatingResultMetric struct {
	MarginPct     decimal.Decimal `json:"margin_pct"`
	ResultUSD     decimal.Decimal `json:"result_usd"`
	TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
}

// ManagementBalance representa el balance de gestión
type ManagementBalance struct {
	Totals BalanceTotals `json:"totals"`
	Items  []BalanceItem `json:"items"`
}

// BalanceSummary representa el resumen del balance
type BalanceSummary struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	StockUSD               decimal.Decimal `json:"stock_usd"`
	RentUSD                decimal.Decimal `json:"rent_usd"`
	StructureUSD           decimal.Decimal `json:"structure_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
}

// BalanceItem representa un item del balance de gestión
type BalanceItem struct {
	Category    BalanceCategory  `json:"category"`
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd,omitempty"`
	Difference  *decimal.Decimal `json:"difference_usd,omitempty"`
	Order       int              `json:"order"`
}

// BalanceBreakdown representa el desglose por categoría (mantenido para compatibilidad)
type BalanceBreakdown struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
}

// BalanceTotals representa la fila de totales
type BalanceTotals struct {
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	InvestedUSD decimal.Decimal `json:"invested_usd"`
	StockUSD    decimal.Decimal `json:"stock_usd"`
}

// CropIncidence representa la incidencia por cultivo
type CropIncidence struct {
	Items []CropItem `json:"items"`
	Total CropTotal  `json:"total"`
}

// CropItem representa un item de cultivo
type CropItem struct {
	CropID       int64           `json:"crop_id"`
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	CostPerHaUSD decimal.Decimal `json:"cost_per_ha_usd"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// CropData representa los datos de un cultivo específico (mantenido para compatibilidad)
type CropData struct {
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// CropTotal representa los totales de cultivos
type CropTotal struct {
	Hectares        decimal.Decimal `json:"hectares"`
	AvgCostPerHaUSD decimal.Decimal `json:"avg_cost_per_ha_usd"`
}

// OperationalIndicators representa los indicadores operativos
type OperationalIndicators struct {
	Items []OperationalItem `json:"items"`
}

// OperationalItem representa un item de indicador operativo
type OperationalItem struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Date        *string `json:"date"`
	WorkorderID *string `json:"workorder_id,omitempty"`
}

// OperationalCard representa una tarjeta de indicador operativo (mantenido para compatibilidad)
type OperationalCard struct {
	Key         string     `json:"key"`
	Title       string     `json:"title"`
	Date        *time.Time `json:"date"`
	WorkorderID *int64     `json:"workorder_id"`
}

// FromDashboardPayloadResponse convierte directamente la respuesta de la función SQL a DTO
func FromDashboardPayloadResponse(payload any) DashboardResponse {
	// Convertir el payload a JSON y luego parsearlo
	jsonData, err := json.Marshal(payload)
	if err != nil {
		// Si hay error, retornar estructura vacía con valores por defecto
		return createEmptyDashboardResponse()
	}

	var response DashboardResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		// Si hay error, retornar estructura vacía con valores por defecto
		return createEmptyDashboardResponse()
	}

	// Aplicar redondeo al entero más cercano a todos los campos decimal
	response = RoundAllDecimals(response)

	return response
}

// createEmptyDashboardResponse crea una respuesta vacía con valores por defecto
func createEmptyDashboardResponse() DashboardResponse {
	return DashboardResponse{
		SchemaVersion: "1.0.0",
		Metrics: Metrics{
			Sowing: SowingMetric{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Harvest: HarvestMetric{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Costs: CostsMetric{
				ProgressPct: decimal.Zero,
				ExecutedUSD: decimal.Zero,
				BudgetUSD:   decimal.Zero,
			},
			InvestorContributions: InvestorContributions{
				Items: []InvestorItem{},
			},
			OperatingResult: OperatingResultMetric{
				MarginPct:     decimal.Zero,
				ResultUSD:     decimal.Zero,
				TotalCostsUSD: decimal.Zero,
			},
		},
		ManagementBalance: ManagementBalance{
			Totals: BalanceTotals{
				ExecutedUSD: decimal.Zero,
				InvestedUSD: decimal.Zero,
				StockUSD:    decimal.Zero,
			},
			Items: []BalanceItem{
				{
					Category:    BalanceCategoryDirectCosts,
					Label:       "Costos Directos",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					StockUSD:    &decimal.Zero,
					Order:       0,
				},
				{
					Category:    BalanceCategorySeed,
					Label:       "Semilla",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					Order:       1,
				},
				{
					Category:    BalanceCategorySupplies,
					Label:       "Agroquímicos",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					Order:       2,
				},
				{
					Category:    BalanceCategoryFertilizers,
					Label:       "Fertilizantes",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					Order:       3,
				},
				{
					Category:    BalanceCategoryLabors,
					Label:       "Labores",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					StockUSD:    &decimal.Zero,
					Order:       4,
				},
				{
					Category:    BalanceCategoryLease,
					Label:       "Arriendo",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					Order:       5,
				},
				{
					Category:    BalanceCategoryAdmin,
					Label:       "Estructura",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					Order:       6,
				},
			},
		},
		CropIncidence: CropIncidence{
			Items: []CropItem{},
			Total: CropTotal{
				Hectares:        decimal.Zero,
				AvgCostPerHaUSD: decimal.Zero,
			},
		},
		OperationalIndicators: OperationalIndicators{
			Items: []OperationalItem{},
		},
	}
}

// FromDashboard convierte el dominio Dashboard a DTO
func FromDashboard(domain *domain.Dashboard) DashboardResponse {
	// Convertir el dominio a JSON y luego parsearlo
	jsonData, err := json.Marshal(domain)
	if err != nil {
		// Si hay error, retornar estructura vacía
		return DashboardResponse{}
	}

	var response DashboardResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		// Si hay error, retornar estructura vacía
		return DashboardResponse{}
	}

	// Aplicar redondeo al entero más cercano a todos los campos decimal
	response = RoundAllDecimals(response)

	return response
}

// Funciones auxiliares para crear valores decimal y punteros
func DecimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func DecimalZero() decimal.Decimal {
	return decimal.Zero
}

func StringPtr(s string) *string {
	return &s
}

// RoundAllDecimals aplica redondeo al entero más cercano a todos los campos decimal
func RoundAllDecimals(response DashboardResponse) DashboardResponse {
	// Redondear métricas
	response.Metrics.Sowing.ProgressPct = roundToNearestInteger(response.Metrics.Sowing.ProgressPct)
	response.Metrics.Sowing.Hectares = roundToNearestInteger(response.Metrics.Sowing.Hectares)
	response.Metrics.Sowing.TotalHectares = roundToNearestInteger(response.Metrics.Sowing.TotalHectares)

	response.Metrics.Harvest.ProgressPct = roundToNearestInteger(response.Metrics.Harvest.ProgressPct)
	response.Metrics.Harvest.Hectares = roundToNearestInteger(response.Metrics.Harvest.Hectares)
	response.Metrics.Harvest.TotalHectares = roundToNearestInteger(response.Metrics.Harvest.TotalHectares)

	response.Metrics.Costs.ProgressPct = roundToNearestInteger(response.Metrics.Costs.ProgressPct)
	response.Metrics.Costs.ExecutedUSD = roundToNearestInteger(response.Metrics.Costs.ExecutedUSD)
	response.Metrics.Costs.BudgetUSD = roundToNearestInteger(response.Metrics.Costs.BudgetUSD)

	// Redondear contribuciones de inversores
	for i := range response.Metrics.InvestorContributions.Items {
		response.Metrics.InvestorContributions.Items[i].SharePct = roundToNearestInteger(response.Metrics.InvestorContributions.Items[i].SharePct)
		// Redondear contributions_progress_pct a 2 decimales para mostrar como "75.74%"
		response.Metrics.InvestorContributions.Items[i].ContributionsProgressPct = response.Metrics.InvestorContributions.Items[i].ContributionsProgressPct.Round(2)
	}

	response.Metrics.OperatingResult.MarginPct = roundToNearestInteger(response.Metrics.OperatingResult.MarginPct)
	response.Metrics.OperatingResult.ResultUSD = roundToNearestInteger(response.Metrics.OperatingResult.ResultUSD)
	response.Metrics.OperatingResult.TotalCostsUSD = roundToNearestInteger(response.Metrics.OperatingResult.TotalCostsUSD)

	// Redondear totales del balance de gestión
	response.ManagementBalance.Totals.ExecutedUSD = roundToNearestInteger(response.ManagementBalance.Totals.ExecutedUSD)
	response.ManagementBalance.Totals.InvestedUSD = roundToNearestInteger(response.ManagementBalance.Totals.InvestedUSD)
	response.ManagementBalance.Totals.StockUSD = roundToNearestInteger(response.ManagementBalance.Totals.StockUSD)

	// Redondear items del balance de gestión
	for i := range response.ManagementBalance.Items {
		response.ManagementBalance.Items[i].ExecutedUSD = roundToNearestInteger(response.ManagementBalance.Items[i].ExecutedUSD)
		response.ManagementBalance.Items[i].InvestedUSD = roundToNearestInteger(response.ManagementBalance.Items[i].InvestedUSD)
		if response.ManagementBalance.Items[i].StockUSD != nil {
			*response.ManagementBalance.Items[i].StockUSD = roundToNearestInteger(*response.ManagementBalance.Items[i].StockUSD)
		}
		if response.ManagementBalance.Items[i].Difference != nil {
			*response.ManagementBalance.Items[i].Difference = roundToNearestInteger(*response.ManagementBalance.Items[i].Difference)
		}
	}

	// Redondear incidencia por cultivo
	for i := range response.CropIncidence.Items {
		response.CropIncidence.Items[i].Hectares = roundToNearestInteger(response.CropIncidence.Items[i].Hectares)
		response.CropIncidence.Items[i].CostPerHaUSD = roundToNearestInteger(response.CropIncidence.Items[i].CostPerHaUSD)
		response.CropIncidence.Items[i].IncidencePct = roundToNearestInteger(response.CropIncidence.Items[i].IncidencePct)
	}

	response.CropIncidence.Total.Hectares = roundToNearestInteger(response.CropIncidence.Total.Hectares)
	response.CropIncidence.Total.AvgCostPerHaUSD = roundToNearestInteger(response.CropIncidence.Total.AvgCostPerHaUSD)

	return response
}

// ===== MAPPER FUNCTIONS =====

// ToDashboardFilter convierte un DTO de filtro a entidad de dominio
func ToDashboardFilter(dto DashboardFilterRequest) domain.DashboardFilter {
	return domain.DashboardFilter{
		CustomerID: dto.CustomerID,
		ProjectID:  dto.ProjectID,
		CampaignID: dto.CampaignID,
		FieldID:    dto.FieldID,
	}
}

// FromDashboardData convierte el dominio DashboardData a DTO
func FromDashboardData(domain *domain.DashboardData) DashboardResponse {
	if domain == nil {
		return createEmptyDashboardResponse()
	}

	response := DashboardResponse{
		SchemaVersion:         "1.0.0",
		Metrics:               convertMetrics(domain.Metrics),
		ManagementBalance:     convertManagementBalance(domain.ManagementBalance),
		CropIncidence:         convertCropIncidence(domain.CropIncidence),
		OperationalIndicators: convertOperationalIndicators(domain.OperationalIndicators),
	}

	// Aplicar redondeo al entero más cercano a todos los campos decimal
	response = RoundAllDecimals(response)

	return response
}

// convertMetrics convierte las métricas del dominio al DTO
func convertMetrics(metrics *domain.DashboardMetrics) Metrics {
	if metrics == nil {
		return Metrics{}
	}

	return Metrics{
		Sowing:                convertSowing(metrics.Sowing),
		Harvest:               convertHarvest(metrics.Harvest),
		Costs:                 convertCosts(metrics.Costs),
		InvestorContributions: convertInvestorContributions(metrics.InvestorContributions),
		OperatingResult:       convertOperatingResult(metrics.OperatingResult),
	}
}

// convertSowing convierte la métrica de siembra
func convertSowing(sowing *domain.DashboardSowing) SowingMetric {
	if sowing == nil {
		return SowingMetric{}
	}

	return SowingMetric{
		ProgressPct:   sowing.ProgressPct,
		Hectares:      sowing.Hectares,
		TotalHectares: sowing.TotalHectares,
	}
}

// convertHarvest convierte la métrica de cosecha
func convertHarvest(harvest *domain.DashboardHarvest) HarvestMetric {
	if harvest == nil {
		return HarvestMetric{}
	}

	return HarvestMetric{
		ProgressPct:   harvest.ProgressPct,
		Hectares:      harvest.Hectares,
		TotalHectares: harvest.TotalHectares,
	}
}

// convertCosts convierte la métrica de costos
func convertCosts(costs *domain.DashboardCosts) CostsMetric {
	if costs == nil {
		return CostsMetric{}
	}

	return CostsMetric{
		ProgressPct: costs.ProgressPct,
		ExecutedUSD: costs.ExecutedUSD,
		BudgetUSD:   costs.BudgetUSD,
	}
}

// convertInvestorContributions convierte las contribuciones de inversores
func convertInvestorContributions(contributions *domain.DashboardInvestorContributions) InvestorContributions {
	if contributions == nil {
		return InvestorContributions{}
	}

	items := make([]InvestorItem, 0, len(contributions.Breakdown))
	for _, inv := range contributions.Breakdown {
		items = append(items, InvestorItem{
			InvestorID:               inv.InvestorID,
			InvestorName:             inv.InvestorName,
			SharePct:                 inv.PercentPct,
			ContributionsProgressPct: inv.ContributionsProgressPct,
		})
	}

	return InvestorContributions{
		Items: items,
	}
}

// convertOperatingResult convierte la métrica de resultado operativo
func convertOperatingResult(result *domain.DashboardOperatingResult) OperatingResultMetric {
	if result == nil {
		return OperatingResultMetric{}
	}

	return OperatingResultMetric{
		MarginPct:     result.ProgressPct, // Usar ProgressPct como MarginPct
		ResultUSD:     result.ResultUSD,
		TotalCostsUSD: result.TotalCostsUSD,
	}
}

// convertManagementBalance convierte el balance de gestión
func convertManagementBalance(balance *domain.DashboardManagementBalance) ManagementBalance {
	if balance == nil {
		return ManagementBalance{}
	}

	decimalPtr := func(value decimal.Decimal) *decimal.Decimal {
		v := value
		return &v
	}

	leaseExecutedUSD := balance.Summary.RentExecutedUSD
	leaseInvestedUSD := balance.Summary.RentUSD
	if leaseInvestedUSD.GreaterThan(leaseExecutedUSD) {
		leaseExecutedUSD, leaseInvestedUSD = leaseInvestedUSD, leaseExecutedUSD
	}

	// Crear totales usando SOLO datos de BD (sin hardcoding)
	totals := BalanceTotals{
		ExecutedUSD: balance.Summary.DirectCostsExecutedUSD,
		InvestedUSD: balance.Summary.DirectCostsInvestedUSD,
		StockUSD:    balance.Summary.StockUSD,
	}

	// Crear items con todos los campos del balance
	items := []BalanceItem{
		{
			Category:    BalanceCategoryDirectCosts,
			Label:       "Costos Directos",
			ExecutedUSD: balance.Summary.DirectCostsExecutedUSD,
			InvestedUSD: balance.Summary.DirectCostsInvestedUSD,
			StockUSD:    &balance.Summary.StockUSD,
			Difference:  decimalPtr(balance.Summary.DirectCostsExecutedUSD.Sub(balance.Summary.DirectCostsInvestedUSD)),
			Order:       0,
		},
		{
			Category:    BalanceCategorySeed,
			Label:       "Semilla",
			ExecutedUSD: balance.Summary.SemillaCostUSD,
			InvestedUSD: balance.Summary.SemillasInvertidosUSD,
			StockUSD:    &balance.Summary.SemillasStockUSD,
			Order:       1,
		},
		{
			Category:    BalanceCategorySupplies,
			Label:       "Agroquímicos",
			ExecutedUSD: balance.Summary.InsumosCostUSD,
			InvestedUSD: balance.Summary.AgroquimicosInvertidosUSD,
			StockUSD:    &balance.Summary.AgroquimicosStockUSD,
			Order:       2,
		},
		{
			Category:    BalanceCategoryFertilizers,
			Label:       "Fertilizantes",
			ExecutedUSD: balance.Summary.FertilizantesCostUSD,
			InvestedUSD: balance.Summary.FertilizantesInvertidosUSD,
			StockUSD:    &balance.Summary.FertilizantesStockUSD,
			Order:       3,
		},
		{
			Category:    BalanceCategoryLabors,
			Label:       "Labores",
			ExecutedUSD: balance.Summary.LaboresCostUSD,
			InvestedUSD: balance.Summary.LaboresInvertidosUSD,
			StockUSD:    &decimal.Zero,
			Order:       4,
		},
		{
			Category:    BalanceCategoryLease,
			Label:       "Arriendo",
			ExecutedUSD: leaseExecutedUSD,
			InvestedUSD: leaseInvestedUSD,
			Difference:  decimalPtr(leaseExecutedUSD.Sub(leaseInvestedUSD)),
			Order:       5,
		},
		{
			Category:    BalanceCategoryAdmin,
			Label:       "Estructura",
			ExecutedUSD: balance.Summary.StructureExecutedUSD,
			InvestedUSD: balance.Summary.StructureUSD,
			Order:       6,
		},
	}

	return ManagementBalance{
		Totals: totals,
		Items:  items,
	}
}

// convertBalanceSummary convierte el resumen del balance

// convertBalanceBreakdown convierte el desglose del balance

// convertCropIncidence convierte la incidencia de cultivos
func convertCropIncidence(incidence *domain.DashboardCropIncidence) CropIncidence {
	if incidence == nil {
		return CropIncidence{}
	}

	items := make([]CropItem, 0, len(incidence.Crops))
	for _, crop := range incidence.Crops {
		items = append(items, CropItem{
			CropID:       crop.ID,
			Name:         crop.Name,
			Hectares:     crop.Hectares,
			CostPerHaUSD: crop.CostUSDPerHa,
			IncidencePct: crop.IncidencePct,
		})
	}

	// Calcular totales agregando los datos de los items
	var totalHectares, totalCostUSD decimal.Decimal
	for _, item := range items {
		totalHectares = totalHectares.Add(item.Hectares)
		// Costo total del cultivo = costo por hectárea × hectáreas del cultivo
		cropTotalCost := item.CostPerHaUSD.Mul(item.Hectares)
		totalCostUSD = totalCostUSD.Add(cropTotalCost)
	}

	// Calcular costo promedio por hectárea: costo total / hectáreas totales
	var avgCostPerHaUSD decimal.Decimal
	if !totalHectares.IsZero() {
		avgCostPerHaUSD = totalCostUSD.Div(totalHectares)
	}

	// Crear el total calculado
	calculatedTotal := CropTotal{
		Hectares:        totalHectares,
		AvgCostPerHaUSD: avgCostPerHaUSD,
	}

	return CropIncidence{
		Items: items,
		Total: calculatedTotal,
	}
}

// convertOperationalIndicators convierte los indicadores operativos
func convertOperationalIndicators(indicators *domain.DashboardOperationalIndicators) OperationalIndicators {
	if indicators == nil {
		return OperationalIndicators{}
	}

	items := make([]OperationalItem, 0, len(indicators.Cards))
	for _, card := range indicators.Cards {
		// Convertir time.Time a string si es necesario
		var dateStr *string
		if card.Date != nil {
			dateStrVal := card.Date.Format("2006-01-02")
			dateStr = &dateStrVal
		}

		// Solo incluir workorder_id para tipos que lo requieren
		var workorderID *string
		if card.WorkorderID != nil && (card.Key == "first_workorder" || card.Key == "last_workorder") {
			workorderID = card.WorkorderID
		}

		items = append(items, OperationalItem{
			Type:        card.Key, // Usar Key como Type
			Title:       card.Title,
			Date:        dateStr,
			WorkorderID: workorderID,
		})
	}

	return OperationalIndicators{
		Items: items,
	}
}

// ToDashboard convierte un DTO de dashboard a entidad de dominio
// (útil para casos donde se recibe un dashboard desde el exterior)
func ToDashboard(dto DashboardRequest) *domain.Dashboard {
	return &domain.Dashboard{
		// Campos básicos
	}
}
