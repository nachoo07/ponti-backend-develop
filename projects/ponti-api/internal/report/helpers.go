// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// ===== TIPOS REUTILIZABLES =====

// MetricExtractor es una función que extrae un valor decimal de una métrica
type MetricExtractor func(domain.FieldCropMetric) decimal.Decimal

// RowConfig define la configuración para construir una fila del reporte
type RowConfig struct {
	Key       string
	Unit      string
	ValueType string
	Extractor MetricExtractor
}

// ===== EXTRACTORES COMUNES (DRY) =====

// Extractores de información básica
var (
	ExtractSurface    MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.SurfaceHa }
	ExtractProduction MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.ProductionTn }
	ExtractYield      MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.YieldTnHa }
)

// Extractores de precios y comercialización
var (
	ExtractFreightCost    MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.FreightCostUsdTn }
	ExtractCommercialCost MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.CommercialCostUsdTn }
	ExtractNetPrice       MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.NetPriceUsdTn }
	ExtractGrossPrice     MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.GrossPriceUsdTn }
	ExtractNetIncome      MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.NetIncomeUsdHa }
)

// Extractores de costos directos
var (
	ExtractLaborCosts  MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.LaborCostsUsdHa }
	ExtractSupplyCosts MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.SupplyCostsUsdHa }
	ExtractDirectCosts MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.DirectCostsUsdHa }
	ExtractGrossMargin MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.GrossMarginUsdHa }
)

// Extractores de costos adicionales
var (
	ExtractRent            MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.RentUsdHa }
	ExtractAdministration  MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.AdministrationUsdHa }
	ExtractOperatingResult MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.OperatingResultUsdHa }
)

// Extractores de métricas adicionales
var (
	ExtractTotalInvested     MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.TotalInvestedUsdHa }
	ExtractReturnPct         MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.ReturnPct }
	ExtractIndifferenceYield MetricExtractor = func(m domain.FieldCropMetric) decimal.Decimal { return m.IndifferenceYieldUsdTn }
)

// ===== EXTRACTORES CALCULADOS =====

// ExtractIndifferencePrice calcula el precio NETO de indiferencia
// Precio Indiferencia = Total Invertido por HA / Rendimiento
// A qué precio necesito vender para cubrir mis costos
func ExtractIndifferencePrice(m domain.FieldCropMetric) decimal.Decimal {
	if m.YieldTnHa.GreaterThan(decimal.Zero) {
		return m.TotalInvestedUsdHa.Div(m.YieldTnHa)
	}
	return decimal.Zero
}

// ===== CONFIGURACIONES DE FILAS (DRY) =====

// GetMainRowConfigs retorna las configuraciones para las filas principales del reporte
func GetMainRowConfigs() []RowConfig {
	return []RowConfig{
		// Información básica
		{Key: "surface", Unit: "ha", ValueType: "number", Extractor: ExtractSurface},
		{Key: "production", Unit: "tn", ValueType: "number", Extractor: ExtractProduction},
		{Key: "yield", Unit: "tn/ha", ValueType: "number", Extractor: ExtractYield},

		// Precios y comercialización
		{Key: "freight_cost", Unit: "usd/tn", ValueType: "number", Extractor: ExtractFreightCost},
		{Key: "commercial_cost", Unit: "usd/tn", ValueType: "number", Extractor: ExtractCommercialCost},
		{Key: "net_price", Unit: "usd/tn", ValueType: "number", Extractor: ExtractNetPrice},
		{Key: "gross_price", Unit: "usd/tn", ValueType: "number", Extractor: ExtractGrossPrice},
		{Key: "net_income", Unit: "usd/ha", ValueType: "number", Extractor: ExtractNetIncome},

		// Costos directos
		{Key: "labors_cost", Unit: "usd/ha", ValueType: "number", Extractor: ExtractLaborCosts},
		{Key: "supplies_cost", Unit: "usd/ha", ValueType: "number", Extractor: ExtractSupplyCosts},
		{Key: "total_direct_costs", Unit: "usd/ha", ValueType: "number", Extractor: ExtractDirectCosts},
		{Key: "gross_margin", Unit: "usd/ha", ValueType: "number", Extractor: ExtractGrossMargin},

		// Costos adicionales
		{Key: "lease", Unit: "usd/ha", ValueType: "number", Extractor: ExtractRent},
		{Key: "admin", Unit: "usd/ha", ValueType: "number", Extractor: ExtractAdministration},
		{Key: "operating_result", Unit: "usd/ha", ValueType: "number", Extractor: ExtractOperatingResult},

		// Métricas adicionales
		{Key: "total_invested", Unit: "usd/ha", ValueType: "number", Extractor: ExtractTotalInvested},
		{Key: "return_pct", Unit: "%", ValueType: "number", Extractor: ExtractReturnPct},
		{Key: "indifference_yield", Unit: "tn/ha", ValueType: "number", Extractor: ExtractIndifferenceYield},
		{Key: "indifference_price", Unit: "usd/tn", ValueType: "number", Extractor: ExtractIndifferencePrice},
	}
}

// ===== FUNCIONES HELPER (DRY) =====

// BuildRowFromConfig construye una fila del reporte desde una configuración
func BuildRowFromConfig(
	config RowConfig,
	metricMap map[string]domain.FieldCropMetric,
	columnMap map[string]domain.FieldCropColumn,
) domain.FieldCropRow {
	values := make(map[string]domain.FieldCropValue)

	for colID := range columnMap {
		if metric, exists := metricMap[colID]; exists {
			values[colID] = domain.FieldCropValue{
				Number: config.Extractor(metric),
			}
		} else {
			values[colID] = domain.FieldCropValue{
				Number: decimal.Zero,
			}
		}
	}

	return domain.FieldCropRow{
		Key:       config.Key,
		Unit:      config.Unit,
		ValueType: config.ValueType,
		Values:    values,
	}
}

// BuildDetailRow construye una fila de detalle (supplies/labors) de forma genérica
func BuildDetailRow(
	key string,
	unit string,
	columnMap map[string]domain.FieldCropColumn,
	extractor func(colID string) decimal.Decimal,
) domain.FieldCropRow {
	values := make(map[string]domain.FieldCropValue)

	for colID := range columnMap {
		values[colID] = domain.FieldCropValue{
			Number: extractor(colID),
		}
	}

	return domain.FieldCropRow{
		Key:       key,
		Unit:      unit,
		ValueType: "number",
		Values:    values,
	}
}
