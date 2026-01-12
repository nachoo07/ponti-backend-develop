// Package models contiene helpers comunes para mappers
package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== HELPERS PARA AGREGACIÓN DE MÉTRICAS (DRY) =====

// MetricAggregator encapsula la lógica de agregación de métricas
type MetricAggregator struct{}

// NewMetricAggregator crea una nueva instancia
func NewMetricAggregator() *MetricAggregator {
	return &MetricAggregator{}
}

// SumMetrics suma todas las métricas en un solo struct
func (agg *MetricAggregator) SumMetrics(metrics []*domain.FieldCropMetric) domain.FieldCropMetric {
	total := domain.FieldCropMetric{}

	for _, metric := range metrics {
		total.SurfaceHa = total.SurfaceHa.Add(metric.SurfaceHa)
		total.ProductionTn = total.ProductionTn.Add(metric.ProductionTn)
		total.SownAreaHa = total.SownAreaHa.Add(metric.SownAreaHa)
		total.HarvestedAreaHa = total.HarvestedAreaHa.Add(metric.HarvestedAreaHa)
		total.NetIncomeUsd = total.NetIncomeUsd.Add(metric.NetIncomeUsd)
		total.LaborCostsUsd = total.LaborCostsUsd.Add(metric.LaborCostsUsd)
		total.SupplyCostsUsd = total.SupplyCostsUsd.Add(metric.SupplyCostsUsd)
		total.TotalDirectCostsUsd = total.TotalDirectCostsUsd.Add(metric.TotalDirectCostsUsd)
		total.GrossMarginUsd = total.GrossMarginUsd.Add(metric.GrossMarginUsd)
		total.RentUsd = total.RentUsd.Add(metric.RentUsd)
		total.AdministrationUsd = total.AdministrationUsd.Add(metric.AdministrationUsd)
		total.OperatingResultUsd = total.OperatingResultUsd.Add(metric.OperatingResultUsd)
		total.TotalInvestedUsd = total.TotalInvestedUsd.Add(metric.TotalInvestedUsd)
	}

	return total
}

// CalculateRatios calcula todos los ratios y promedios de una métrica
func (agg *MetricAggregator) CalculateRatios(metric *domain.FieldCropMetric) {
	// Calcular rendimiento
	if metric.HarvestedAreaHa.GreaterThan(decimal.Zero) {
		metric.YieldTnHa = metric.ProductionTn.Div(metric.HarvestedAreaHa)
	}

	// Calcular métricas por hectárea
	if metric.SownAreaHa.GreaterThan(decimal.Zero) {
		metric.NetIncomeUsdHa = metric.NetIncomeUsd.Div(metric.SownAreaHa)
		metric.DirectCostsUsdHa = metric.TotalDirectCostsUsd.Div(metric.SownAreaHa)
		metric.GrossMarginUsdHa = metric.GrossMarginUsd.Div(metric.SownAreaHa)
		metric.RentUsdHa = metric.RentUsd.Div(metric.SownAreaHa)
		metric.AdministrationUsdHa = metric.AdministrationUsd.Div(metric.SownAreaHa)
		metric.OperatingResultUsdHa = metric.OperatingResultUsd.Div(metric.SownAreaHa)
		metric.TotalInvestedUsdHa = metric.TotalInvestedUsd.Div(metric.SownAreaHa)
	}

	// Calcular rentabilidad
	if metric.TotalInvestedUsd.GreaterThan(decimal.Zero) {
		metric.ReturnPct = metric.OperatingResultUsd.Div(metric.TotalInvestedUsd)
	}

	// Calcular rinde de indiferencia
	if metric.YieldTnHa.GreaterThan(decimal.Zero) {
		metric.IndifferenceYieldUsdTn = metric.TotalInvestedUsd.Div(metric.YieldTnHa)
	}
}

// CopyFirstNonZeroPrice copia el primer precio no cero de las métricas
func (agg *MetricAggregator) CopyFirstNonZeroPrice(metrics []*domain.FieldCropMetric, target *domain.FieldCropMetric) {
	// Precio bruto
	for _, metric := range metrics {
		if metric.GrossPriceUsdTn.GreaterThan(decimal.Zero) {
			target.GrossPriceUsdTn = metric.GrossPriceUsdTn
			break
		}
	}

	// Flete
	for _, metric := range metrics {
		if metric.FreightCostUsdTn.GreaterThan(decimal.Zero) {
			target.FreightCostUsdTn = metric.FreightCostUsdTn
			break
		}
	}

	// Gasto comercial
	for _, metric := range metrics {
		if metric.CommercialCostUsdTn.GreaterThan(decimal.Zero) {
			target.CommercialCostUsdTn = metric.CommercialCostUsdTn
			break
		}
	}

	// Precio neto
	for _, metric := range metrics {
		if metric.NetPriceUsdTn.GreaterThan(decimal.Zero) {
			target.NetPriceUsdTn = metric.NetPriceUsdTn
			break
		}
	}
}

// ===== HELPERS PARA CONVERSIÓN DE SLICES (DRY) =====

// ConvertInvestorSharesSlice convierte un slice de modelos a dominio
func ConvertInvestorSharesSlice(shares []InvestorShareModel) []domain.InvestorShare {
	domainShares := make([]domain.InvestorShare, len(shares))
	for i, s := range shares {
		domainShares[i] = domain.InvestorShare{
			InvestorRef: domain.InvestorRef{
				InvestorID:   s.InvestorID,
				InvestorName: s.InvestorName,
			},
			AmountUsd: s.AmountUsd,
			SharePct:  s.SharePct,
		}
	}
	return domainShares
}
