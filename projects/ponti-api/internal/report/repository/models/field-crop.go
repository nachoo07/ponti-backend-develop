// Package models contiene los modelos de base de datos para los reportes
package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	shareddb "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/db"
	"github.com/shopspring/decimal"
)

// ===== MODELOS DE BASE DE DATOS =====

// FieldCropMetricModel representa el modelo de base de datos para métricas por campo y cultivo
type FieldCropMetricModel struct {
	ProjectID int64  `gorm:"column:project_id"`
	FieldID   int64  `gorm:"column:field_id"`
	FieldName string `gorm:"column:field_name"`
	CropID    int64  `gorm:"column:current_crop_id"`
	CropName  string `gorm:"column:crop_name"`

	// Información general
	SurfaceHa       decimal.Decimal `gorm:"column:superficie_ha"`
	ProductionTn    decimal.Decimal `gorm:"column:produccion_tn"`
	SownAreaHa      decimal.Decimal `gorm:"column:area_sembrada_ha"`
	HarvestedAreaHa decimal.Decimal `gorm:"column:area_cosechada_ha"`

	// Rendimiento
	YieldTnHa decimal.Decimal `gorm:"column:rendimiento_tn_ha"`

	// Precios y comercialización
	GrossPriceUsdTn     decimal.Decimal `gorm:"column:precio_bruto_usd_tn"`
	FreightCostUsdTn    decimal.Decimal `gorm:"column:gasto_flete_usd_tn"`
	CommercialCostUsdTn decimal.Decimal `gorm:"column:gasto_comercial_usd_tn"`
	NetPriceUsdTn       decimal.Decimal `gorm:"column:precio_neto_usd_tn"`

	// Ingreso neto
	NetIncomeUsd   decimal.Decimal `gorm:"column:ingreso_neto_usd"`
	NetIncomeUsdHa decimal.Decimal `gorm:"column:ingreso_neto_usd_ha"`

	// Costos directos
	LaborCostsUsd       decimal.Decimal `gorm:"column:costos_labores_usd"`
	LaborCostsUsdHa     decimal.Decimal `gorm:"column:costos_labores_usd_ha"` // TODO: Agregar a vista SQL v3
	SupplyCostsUsd      decimal.Decimal `gorm:"column:costos_insumos_usd"`
	SupplyCostsUsdHa    decimal.Decimal `gorm:"column:costos_insumos_usd_ha"` // TODO: Agregar a vista SQL v3
	TotalDirectCostsUsd decimal.Decimal `gorm:"column:total_costos_directos_usd"`
	DirectCostsUsdHa    decimal.Decimal `gorm:"column:costos_directos_usd_ha"`

	// Margen bruto
	GrossMarginUsd   decimal.Decimal `gorm:"column:margen_bruto_usd"`
	GrossMarginUsdHa decimal.Decimal `gorm:"column:margen_bruto_usd_ha"`

	// Arriendo
	RentUsd   decimal.Decimal `gorm:"column:arriendo_usd"`
	RentUsdHa decimal.Decimal `gorm:"column:arriendo_usd_ha"`

	// Costos administrativos
	AdministrationUsd   decimal.Decimal `gorm:"column:administracion_usd"`
	AdministrationUsdHa decimal.Decimal `gorm:"column:administracion_usd_ha"`

	// Resultado operativo
	OperatingResultUsd   decimal.Decimal `gorm:"column:resultado_operativo_usd"`
	OperatingResultUsdHa decimal.Decimal `gorm:"column:resultado_operativo_usd_ha"`

	// Total invertido
	TotalInvestedUsd   decimal.Decimal `gorm:"column:total_invertido_usd"`
	TotalInvestedUsdHa decimal.Decimal `gorm:"column:total_invertido_usd_ha"`

	// Métricas calculadas
	ReturnPct              decimal.Decimal `gorm:"column:renta_pct"`
	IndifferenceYieldUsdTn decimal.Decimal `gorm:"column:rinde_indiferencia_usd_tn"`
}

// TableName especifica el nombre de la tabla para GORM
// ACTUALIZADO: Usar helper para switch v3/v4 - Migración 000316
func (FieldCropMetricModel) TableName() string {
	return shareddb.FieldCropView("metrics")
}

// LaborMetricModel representa el modelo de base de datos para métricas de labores
type LaborMetricModel struct {
	ProjectID      int64           `gorm:"column:project_id"`
	FieldID        int64           `gorm:"column:field_id"`
	CropID         int64           `gorm:"column:crop_id"`
	CropName       string          `gorm:"column:crop_name"`
	LaborID        int64           `gorm:"column:labor_id"`
	LaborName      string          `gorm:"column:labor_name"`
	CategoryID     int64           `gorm:"column:category_id"`
	CategoryName   string          `gorm:"column:category_name"`
	SurfaceHa      decimal.Decimal `gorm:"column:surface_ha"`
	CostUsd        decimal.Decimal `gorm:"column:cost_usd"`
	CostPerHa      decimal.Decimal `gorm:"column:cost_per_ha"`
	WorkOrderCount int64           `gorm:"column:workorder_count"`
}

// SupplyMetricModel representa el modelo de base de datos para métricas de insumos
type SupplyMetricModel struct {
	ProjectID      int64           `gorm:"column:project_id"`
	FieldID        int64           `gorm:"column:field_id"`
	CropID         int64           `gorm:"column:crop_id"`
	CropName       string          `gorm:"column:crop_name"`
	SupplyID       int64           `gorm:"column:supply_id"`
	SupplyName     string          `gorm:"column:supply_name"`
	CategoryID     int64           `gorm:"column:category_id"`
	CategoryName   string          `gorm:"column:category_name"`
	SurfaceHa      decimal.Decimal `gorm:"column:surface_ha"`
	QuantityUsed   decimal.Decimal `gorm:"column:quantity_used"`
	CostUsd        decimal.Decimal `gorm:"column:cost_usd"`
	CostPerHa      decimal.Decimal `gorm:"column:cost_per_ha"`
	WorkOrderCount int64           `gorm:"column:workorder_count"`
}

// FieldCropLaborDetailModel representa el modelo de la vista v3_report_field_crop_labores
// Migración 000130
type FieldCropLaborDetailModel struct {
	ProjectID int64 `gorm:"column:project_id"`
	FieldID   int64 `gorm:"column:field_id"`
	CropID    int64 `gorm:"column:current_crop_id"`

	// Desglose por categoría de labor (USD/ha)
	SiembraUsdHa       decimal.Decimal `gorm:"column:siembra_usd_ha"`
	PulverizacionUsdHa decimal.Decimal `gorm:"column:pulverizacion_usd_ha"`
	RiegoUsdHa         decimal.Decimal `gorm:"column:riego_usd_ha"`
	CosechaUsdHa       decimal.Decimal `gorm:"column:cosecha_usd_ha"`
	OtrasLaboresUsdHa  decimal.Decimal `gorm:"column:otras_labores_usd_ha"`
	TotalLaboresUsdHa  decimal.Decimal `gorm:"column:total_labores_usd_ha"`
}

// TableName especifica el nombre de la tabla para GORM
func (FieldCropLaborDetailModel) TableName() string {
	return shareddb.FieldCropView("labores")
}

// FieldCropSupplyDetailModel representa el modelo de la vista v3_report_field_crop_insumos
// Migración 000131 (actualizada con Fertilizantes y Otros Insumos)
type FieldCropSupplyDetailModel struct {
	ProjectID int64 `gorm:"column:project_id"`
	FieldID   int64 `gorm:"column:field_id"`
	CropID    int64 `gorm:"column:current_crop_id"`

	// Desglose por categoría de insumo (USD/ha)
	SemillasUsdHa      decimal.Decimal `gorm:"column:semillas_usd_ha"`
	CurasemillasUsdHa  decimal.Decimal `gorm:"column:curasemillas_usd_ha"`
	HerbicidasUsdHa    decimal.Decimal `gorm:"column:herbicidas_usd_ha"`
	InsecticidasUsdHa  decimal.Decimal `gorm:"column:insecticidas_usd_ha"`
	FungicidasUsdHa    decimal.Decimal `gorm:"column:fungicidas_usd_ha"`
	CoadyuvantesUsdHa  decimal.Decimal `gorm:"column:coadyuvantes_usd_ha"`
	FertilizantesUsdHa decimal.Decimal `gorm:"column:fertilizantes_usd_ha"`
	OtrosInsumosUsdHa  decimal.Decimal `gorm:"column:otros_insumos_usd_ha"`
	TotalInsumosUsdHa  decimal.Decimal `gorm:"column:total_insumos_usd_ha"`
}

// TableName especifica el nombre de la tabla para GORM
func (FieldCropSupplyDetailModel) TableName() string {
	return shareddb.FieldCropView("insumos")
}

// ===== MAPPERS =====

// ToDomainFieldCropMetric convierte de modelo a dominio
func (m *FieldCropMetricModel) ToDomainFieldCropMetric() *domain.FieldCropMetric {
	return &domain.FieldCropMetric{
		ProjectID:              m.ProjectID,
		FieldID:                m.FieldID,
		FieldName:              m.FieldName,
		CropID:                 m.CropID,
		CropName:               m.CropName,
		SurfaceHa:              m.SurfaceHa,
		ProductionTn:           m.ProductionTn,
		SownAreaHa:             m.SownAreaHa,
		HarvestedAreaHa:        m.HarvestedAreaHa,
		YieldTnHa:              m.YieldTnHa,
		GrossPriceUsdTn:        m.GrossPriceUsdTn,
		FreightCostUsdTn:       m.FreightCostUsdTn,
		CommercialCostUsdTn:    m.CommercialCostUsdTn,
		NetPriceUsdTn:          m.NetPriceUsdTn,
		NetIncomeUsd:           m.NetIncomeUsd,
		NetIncomeUsdHa:         m.NetIncomeUsdHa,
		LaborCostsUsd:          m.LaborCostsUsd,
		LaborCostsUsdHa:        m.LaborCostsUsdHa, // TODO: Ahora viene de la vista v3
		SupplyCostsUsd:         m.SupplyCostsUsd,
		SupplyCostsUsdHa:       m.SupplyCostsUsdHa, // TODO: Ahora viene de la vista v3
		TotalDirectCostsUsd:    m.TotalDirectCostsUsd,
		DirectCostsUsdHa:       m.DirectCostsUsdHa,
		GrossMarginUsd:         m.GrossMarginUsd,
		GrossMarginUsdHa:       m.GrossMarginUsdHa,
		RentUsd:                m.RentUsd,
		RentUsdHa:              m.RentUsdHa,
		AdministrationUsd:      m.AdministrationUsd,
		AdministrationUsdHa:    m.AdministrationUsdHa,
		OperatingResultUsd:     m.OperatingResultUsd,
		OperatingResultUsdHa:   m.OperatingResultUsdHa,
		TotalInvestedUsd:       m.TotalInvestedUsd,
		TotalInvestedUsdHa:     m.TotalInvestedUsdHa,
		ReturnPct:              m.ReturnPct,
		IndifferenceYieldUsdTn: m.IndifferenceYieldUsdTn,
	}
}

// ToDomainLaborMetric convierte de modelo a dominio
func (m *LaborMetricModel) ToDomainLaborMetric() *domain.LaborMetric {
	return &domain.LaborMetric{
		LaborID:        m.LaborID,
		LaborName:      m.LaborName,
		CategoryID:     m.CategoryID,
		CategoryName:   m.CategoryName,
		SurfaceHa:      m.SurfaceHa,
		CostUsd:        m.CostUsd,
		CostPerHa:      m.CostPerHa,
		WorkOrderCount: m.WorkOrderCount,
	}
}

// ToDomainSupplyMetric convierte de modelo a dominio
func (m *SupplyMetricModel) ToDomainSupplyMetric() *domain.SupplyMetric {
	return &domain.SupplyMetric{
		SupplyID:       m.SupplyID,
		SupplyName:     m.SupplyName,
		CategoryID:     m.CategoryID,
		CategoryName:   m.CategoryName,
		SurfaceHa:      m.SurfaceHa,
		QuantityUsed:   m.QuantityUsed,
		CostUsd:        m.CostUsd,
		CostPerHa:      m.CostPerHa,
		WorkOrderCount: m.WorkOrderCount,
	}
}

// ===== FUNCIONES DE AGRUPACIÓN =====

// GroupFieldCropMetricsByCrop agrupa las métricas por cultivo
func GroupFieldCropMetricsByCrop(metrics []*domain.FieldCropMetric) map[string][]domain.FieldCropMetric {
	grouped := make(map[string][]domain.FieldCropMetric)

	for _, metric := range metrics {
		cropName := metric.CropName
		if cropName == "" {
			cropName = "Sin cultivo"
		}
		grouped[cropName] = append(grouped[cropName], *metric)
	}

	return grouped
}

// CalculateCropTotals calcula los totales por cultivo usando helpers (DRY)
func CalculateCropTotals(metrics []*domain.FieldCropMetric) map[string]domain.FieldCropMetric {
	totals := make(map[string]domain.FieldCropMetric)
	aggregator := NewMetricAggregator()

	// Agrupar por cultivo
	grouped := make(map[string][]*domain.FieldCropMetric)
	for _, metric := range metrics {
		cropName := metric.CropName
		if cropName == "" {
			cropName = "Sin cultivo"
		}
		grouped[cropName] = append(grouped[cropName], metric)
	}

	// Calcular totales para cada cultivo usando helpers
	for cropName, cropMetrics := range grouped {
		if len(cropMetrics) == 0 {
			continue
		}

		// Sumar métricas usando helper
		total := aggregator.SumMetrics(cropMetrics)
		total.CropName = cropName

		// Calcular ratios usando helper
		aggregator.CalculateRatios(&total)

		// Copiar precios usando helper
		aggregator.CopyFirstNonZeroPrice(cropMetrics, &total)

		totals[cropName] = total
	}

	return totals
}

// CalculateGrandTotal calcula el total general usando helpers (DRY)
func CalculateGrandTotal(metrics []*domain.FieldCropMetric) domain.FieldCropMetric {
	aggregator := NewMetricAggregator()

	// Sumar todas las métricas usando helper
	total := aggregator.SumMetrics(metrics)
	total.CropName = "TOTAL GENERAL"

	// Calcular ratios usando helper
	aggregator.CalculateRatios(&total)

	return total
}
