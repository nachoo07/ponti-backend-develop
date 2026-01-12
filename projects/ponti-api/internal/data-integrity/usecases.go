// Package data_integrity implementa casos de uso para validar la coherencia de datos
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
//
// ESTOS CÁLCULOS SON CRÍTICOS Y NO DEBEN ALTERARSE A MENOS QUE SE RECIBA
// UNA ORDEN DIRECTA Y CLARA DEL USUARIO.
//
// REGLAS INVIOLABLES:
// - NUNCA modificar los cálculos LEFT/RIGHT sin autorización explícita
// - NUNCA cambiar las tolerancias sin autorización explícita
// - NUNCA alterar la lógica de los 14 controles sin autorización explícita
// - NUNCA usar ROUND() en cálculos internos (solo en DTOs de salida)
// - SIEMPRE mantener precisión completa en cálculos SQL y Go
//
// Si necesitas modificar algo, DEBES pedir autorización explícita primero.
package data_integrity

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	dashboardDomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/data-integrity/usecases/domain"
	lotDomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	reportDomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	stockDomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	workorderDomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

// WorkorderRepositoryPort define la interfaz para el repositorio de workorders
type WorkorderRepositoryPort interface {
	GetMetrics(ctx context.Context, filt workorderDomain.WorkorderFilter) (*workorderDomain.WorkorderMetrics, error)
	GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error)
}

// DashboardRepositoryPort define la interfaz para el repositorio de dashboard
type DashboardRepositoryPort interface {
	GetDashboard(ctx context.Context, filter dashboardDomain.DashboardFilter) (*dashboardDomain.DashboardData, error)
}

// LotRepositoryPort define la interfaz para el repositorio de lotes
type LotRepositoryPort interface {
	ListLots(ctx context.Context, projectID, fieldID, cropID int64, page, pageSize int) ([]lotDomain.LotTable, int, decimal.Decimal, decimal.Decimal, error)
}

// ReportRepositoryPort define la interfaz para el repositorio de reportes
type ReportRepositoryPort interface {
	GetSummaryResults(filters reportDomain.SummaryResultsFilter) ([]reportDomain.SummaryResults, error)
	GetFieldCropMetrics(filters reportDomain.ReportFilter) ([]reportDomain.FieldCropMetric, error)
	GetInvestorContributionReport(ctx context.Context, filter reportDomain.ReportFilter) (*reportDomain.InvestorContributionReport, error)
}

// StockRepositoryPort define la interfaz para el repositorio de stock
type StockRepositoryPort interface {
	GetStocks(ctx context.Context, projectID int64, closeDate time.Time) ([]*stockDomain.Stock, error)
}

// UseCases contiene los casos de uso del módulo data_integrity
type UseCases struct {
	workorderRepo WorkorderRepositoryPort
	dashboardRepo DashboardRepositoryPort
	lotRepo       LotRepositoryPort
	reportRepo    ReportRepositoryPort
	stockRepo     StockRepositoryPort
}

// NewUseCases crea una nueva instancia de UseCases
func NewUseCases(
	workorderRepo WorkorderRepositoryPort,
	dashboardRepo DashboardRepositoryPort,
	lotRepo LotRepositoryPort,
	reportRepo ReportRepositoryPort,
	stockRepo StockRepositoryPort,
) *UseCases {
	return &UseCases{
		workorderRepo: workorderRepo,
		dashboardRepo: dashboardRepo,
		lotRepo:       lotRepo,
		reportRepo:    reportRepo,
		stockRepo:     stockRepo,
	}
}

// CheckCostsCoherence valida la coherencia de costos con 14 controles individuales
// Cada control calcula LEFT (origen/correcto) y RIGHT (destino/validar) de forma INDEPENDIENTE
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTA FUNCIÓN CONTIENE LOS 14 CONTROLES CRÍTICOS DE INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) CheckCostsCoherence(ctx context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error) {
	checks := make([]domain.IntegrityCheck, 0, 14)

	// GRUPO 1: Costos Directos Ejecutados (Controles 1-4)
	control1, err := u.control1_OrdenesVsDashboard(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 1 failed: %w", err)
	}
	checks = append(checks, control1)

	control2, err := u.control2_OrdenesVsLotes(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 2 failed: %w", err)
	}
	checks = append(checks, control2)

	control3, err := u.control3_OrdenesVsInformeCampo(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 3 failed: %w", err)
	}
	checks = append(checks, control3)

	control4, err := u.control4_OrdenesVsInformeGenerales(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 4 failed: %w", err)
	}
	checks = append(checks, control4)

	// GRUPO 2: Invertidos (Controles 5-7)
	control5, err := u.control5_LaboresInsumosVsDashboard(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 5 failed: %w", err)
	}
	checks = append(checks, control5)

	control6, err := u.control6_LaboresVsAportes(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 6 failed: %w", err)
	}
	checks = append(checks, control6)

	control7, err := u.control7_InsumosVsAportes(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 7 failed: %w", err)
	}
	checks = append(checks, control7)

	// GRUPO 3: Lotes → Aportes (Controles 8-9)
	control8, err := u.control8_LotesAdminVsAportes(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 8 failed: %w", err)
	}
	checks = append(checks, control8)

	control9, err := u.control9_LotesArriendoVsAportes(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 9 failed: %w", err)
	}
	checks = append(checks, control9)

	// GRUPO 4: Ingreso Neto (Control 10)
	control10, err := u.control10_LotesIngresoNetoVsResumen(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 10 failed: %w", err)
	}
	checks = append(checks, control10)

	// GRUPO 5: Resultado Operativo (Controles 11-13)
	control11, err := u.control11_LotesResultadoVsInformeCultivo(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 11 failed: %w", err)
	}
	checks = append(checks, control11)

	control12, err := u.control12_LotesResultadoVsInformeGenerales(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 12 failed: %w", err)
	}
	checks = append(checks, control12)

	control13, err := u.control13_LotesResultadoVsDashboard(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 13 failed: %w", err)
	}
	checks = append(checks, control13)

	// GRUPO 6: Stock (Control 14)
	control14, err := u.control14_StockVsDashboard(ctx, filter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("control 14 failed: %w", err)
	}
	checks = append(checks, control14)

	return &domain.IntegrityReport{
		Checks: checks,
	}, nil
}

// =====================================================
// CONTROL 1: Órdenes de trabajo → Dashboard
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: Dashboard.CostosDirectos (SSOT)
// =====================================================
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTE CONTROL ES CRÍTICO PARA LA INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) control1_OrdenesVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Costos RAW desde workorders
	leftValue, err := u.workorderRepo.GetRawDirectCost(ctx, pID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// RIGHT: Costos desde dashboard (usa funciones SSOT)
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	rightValue := dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD

	return buildCheck(
		1,
		"Órdenes de trabajo",
		"Costos directos ejecutados",
		"Dashboard",
		"Dashboard.CostosDirectos = ∑(Ordenes.costo_total)",
		"∑(workorders.effective_area × labors.price + workorder_items.final_dose × effective_area × supplies.price)",
		leftValue,
		"Tabla workorders RAW",
		"v3_dashboard_ssot.direct_costs_total_for_project()",
		rightValue,
		"Vista v3_dashboard_management_balance",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 2: Órdenes de trabajo → Lotes
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: ∑(Costo_directo_ha_lote × Superficie_lote)
// =====================================================
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTE CONTROL ES CRÍTICO PARA LA INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) control2_OrdenesVsLotes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Costos RAW desde workorders
	leftValue, err := u.workorderRepo.GetRawDirectCost(ctx, pID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// RIGHT: Suma desde lotes (usa SSOT)
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, lot := range lots {
		costTotal := lot.CostUsdPerHa.Mul(lot.SowedArea)
		rightValue = rightValue.Add(costTotal)
	}

	return buildCheck(
		2,
		"Órdenes de trabajo",
		"Costos directos ejecutados",
		"Lotes",
		"∑(Costo_directo_ha_lote × Superficie_lote) = ∑(Órdenes.costo_total)",
		"∑(workorders RAW)",
		leftValue,
		"Tabla workorders RAW",
		"∑(cost_usd_per_ha × sowed_area_ha)",
		rightValue,
		"Vista v3_lot_list",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 3: Órdenes de trabajo → Informe de Resultado por campo
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: ∑(Costo_directo_ha_Cultivo × Superficie_Cultivo)
// =====================================================
func (u *UseCases) control3_OrdenesVsInformeCampo(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Costos RAW desde workorders
	leftValue, err := u.workorderRepo.GetRawDirectCost(ctx, pID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// RIGHT: Desde informe field-crop
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	fieldCropMetrics, err := u.reportRepo.GetFieldCropMetrics(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, metric := range fieldCropMetrics {
		if !metric.TotalDirectCostsUsd.IsZero() {
			rightValue = rightValue.Add(metric.TotalDirectCostsUsd)
		} else {
			costTotal := metric.SurfaceHa.Mul(metric.DirectCostsUsdHa)
			rightValue = rightValue.Add(costTotal)
		}
	}

	return buildCheck(
		3,
		"Órdenes de trabajo",
		"Costos directos ejecutados",
		"Informe de Resultado por campo",
		"∑(Costo_directo_ha_Cultivo × Superficie_Cultivo) = ∑(Órdenes.costo_total)",
		"∑(workorders RAW)",
		leftValue,
		"Tabla workorders RAW",
		"∑(direct_cost_usd por field+crop)",
		rightValue,
		"Vista v3_report_field_crop_metrics",
		decimal.NewFromInt(1), // Tolerancia = 1 USD (diferencias de precisión en agregaciones)
	), nil
}

// =====================================================
// CONTROL 4: Órdenes de trabajo → Informe de Resultado Generales
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: Total de informe generales
// =====================================================
func (u *UseCases) control4_OrdenesVsInformeGenerales(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Costos RAW desde workorders
	leftValue, err := u.workorderRepo.GetRawDirectCost(ctx, pID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// RIGHT: Total del resumen de resultados (primera fila = GRAL CAMPOS)
	filter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	if len(summaryResults) > 0 {
		rightValue = summaryResults[0].TotalDirectCostsUsd
	}

	return buildCheck(
		4,
		"Órdenes de trabajo",
		"Costos directos ejecutados",
		"Informe de Resultado Generales",
		"∑(Costo_directos totales_Cultivo) = ∑(Órdenes.costo_total)",
		"∑(workorders RAW)",
		leftValue,
		"Tabla workorders RAW",
		"summaryResults[0].TotalDirectCostsUsd",
		rightValue,
		"Vista v3_report_summary_results",
		decimal.NewFromInt(1), // Tolerancia = 1 USD (diferencias de precisión en agregaciones)
	), nil
}

// =====================================================
// CONTROL 5: Labores + Insumos → Dashboard
// LEFT: ∑(Labores) + ∑(Insumos) desde dashboard breakdown
// RIGHT: Dashboard.Invertidos total
// =====================================================
func (u *UseCases) control5_LaboresInsumosVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	summary := dashboardData.ManagementBalance.Summary

	// LEFT: Suma de componentes (Labores + Semillas + Agroquímicos + Fertilizantes)
	labores := summary.LaboresInvertidosUSD
	semilla := summary.SemillasInvertidosUSD
	agroquimicos := summary.AgroquimicosInvertidosUSD
	fertilizantes := summary.FertilizantesInvertidosUSD
	leftValue := labores.Add(semilla).Add(agroquimicos).Add(fertilizantes)

	// RIGHT: Total invertidos desde dashboard
	rightValue := dashboardData.ManagementBalance.TotalsRow.InvestedUSD

	return buildCheck(
		5,
		"Labores + Insumos",
		"Invertidos",
		"Dashboard",
		"Dashboard.Invertidos = ∑(Labores) + ∑(Insumos)",
		"Labores + Semillas + Agroquímicos + Fertilizantes",
		leftValue,
		"Dashboard Summary components",
		"TotalsRow.InvestedUSD",
		rightValue,
		"Dashboard TotalsRow",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 6: Labores → Informe de Aportes
// LEFT: Dashboard.LaboresInvertidos (sin cosecha)
// RIGHT: Aportes (Labores Generales + Siembra + Riego)
// =====================================================
func (u *UseCases) control6_LaboresVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// LEFT: Labores invertidas desde dashboard
	leftValue := dashboardData.ManagementBalance.Summary.LaboresInvertidosUSD

	// RIGHT: Informe de aportes (suma de categorías sin cosecha)
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Labores Generales" || category.Label == "Siembra" ||
			category.Label == "Riego" {
			rightValue = rightValue.Add(category.TotalUsd)
		}
	}

	return buildCheck(
		6,
		"Dashboard",
		"Inversión en labores",
		"Informe de Aportes",
		"Dashboard.LaboresInvertidos = ∑(Labores Generales + Siembra + Riego)",
		"Summary.LaboresInvertidosUSD",
		leftValue,
		"Dashboard Summary",
		"∑(Labores Generales + Siembra + Riego)",
		rightValue,
		"Vista investor_contribution_report",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 7: Insumos → Informe de Aportes
// LEFT: Dashboard.InsumosInvertidos
// RIGHT: Aportes (Semillas + Agroquímicos)
// =====================================================
func (u *UseCases) control7_InsumosVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// LEFT: Insumos invertidos desde dashboard
	leftValue := dashboardData.ManagementBalance.Summary.SemillasInvertidosUSD.Add(
		dashboardData.ManagementBalance.Summary.AgroquimicosInvertidosUSD,
	)

	// RIGHT: Informe de aportes
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Semilla" || category.Label == "Agroquímicos" {
			rightValue = rightValue.Add(category.TotalUsd)
		}
	}

	return buildCheck(
		7,
		"Insumos",
		"Inversión en insumos",
		"Informe de Aportes",
		"Dashboard.InsumosInvertidos = ∑(Semilla + Agroquímicos)",
		"Semillas + Agroquímicos",
		leftValue,
		"Dashboard Summary",
		"∑(Semilla + Agroquímicos)",
		rightValue,
		"Vista investor_contribution_report",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 8: Lotes → Informe de Aportes (Administración)
// LEFT: ∑(AdminCost_ha × Superficie) desde lotes
// RIGHT: Aportes Adm.Proyecto
// =====================================================
func (u *UseCases) control8_LotesAdminVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Calcular desde lotes
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		leftValue = leftValue.Add(lot.AdminCost.Mul(lot.SowedArea))
	}

	// RIGHT: Total Aportes Adm.Proyecto del Informe
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Administración y Estructura" {
			rightValue = category.TotalUsd
			break
		}
	}

	return buildCheck(
		8,
		"Lotes",
		"Invertidos",
		"Informe de Aportes",
		"∑(Adm Proyecto_ha_lote × Superficie_lote) = Total Aportes Adm.Proyecto",
		"∑(admin_cost × sowed_area)",
		leftValue,
		"Vista v3_lot_list",
		"Categoría 'Administración y Estructura'",
		rightValue,
		"Vista investor_contribution_report",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 9: Lotes → Informe de Aportes (Arriendo)
// LEFT: ∑(Arriendo_ha × Superficie) desde lotes
// RIGHT: Aportes Arriendo Fijo
// =====================================================
func (u *UseCases) control9_LotesArriendoVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Calcular desde lotes
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		arriendoTotal := lot.RentPerHa.Mul(lot.SowedArea)
		leftValue = leftValue.Add(arriendoTotal)
	}

	// RIGHT: Total Aportes Arriendo del Informe
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Arriendo Capitalizable" {
			rightValue = category.TotalUsd
			break
		}
	}

	return buildCheck(
		9,
		"Lotes",
		"Invertidos",
		"Informe de Aportes",
		"∑(Arriendo Fijo_ha_lote × Superficie_lote) = Total Aportes Arriendo Fijo",
		"∑(rent_per_ha × sowed_area)",
		leftValue,
		"Vista v3_lot_list",
		"Categoría 'Arriendo Capitalizable'",
		rightValue,
		"Vista investor_contribution_report",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 10: Lotes → Resumen de resultados (Ingreso Neto)
// LEFT: ∑(Ingreso_Neto_ha × Superficie) desde lotes
// RIGHT: ∑(Ingreso_Neto) del resumen
// =====================================================
func (u *UseCases) control10_LotesIngresoNetoVsResumen(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Calcular desde lotes
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		ingresoTotal := lot.IncomeNetPerHa.Mul(lot.SowedArea)
		leftValue = leftValue.Add(ingresoTotal)
	}

	// RIGHT: Obtener resumen de resultados
	filter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, result := range summaryResults {
		rightValue = rightValue.Add(result.NetIncomeUsd)
	}

	return buildCheck(
		10,
		"Lotes",
		"Ingreso Neto",
		"Resumen de resultados",
		"∑(Ingreso Neto_ha_lote × Superficie_lote) = ∑(Ingreso Neto del Resumen)",
		"∑(income_net_per_ha × sowed_area)",
		leftValue,
		"Vista v3_lot_list",
		"∑(net_income_usd)",
		rightValue,
		"Vista v3_report_summary_results",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 11: Lotes → Informe por cultivo (Resultado Operativo)
// LEFT: ∑(Resultado.Operativo_ha × Superficie) desde lotes
// RIGHT: ∑(Resultado.Operativo × Superficie) por cultivo
// =====================================================
func (u *UseCases) control11_LotesResultadoVsInformeCultivo(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Calcular desde lotes
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		resultadoTotal := lot.OperatingResultPerHa.Mul(lot.SowedArea)
		leftValue = leftValue.Add(resultadoTotal)
	}

	// RIGHT: Informe por cultivo
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	fieldCropMetrics, err := u.reportRepo.GetFieldCropMetrics(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, metric := range fieldCropMetrics {
		resultadoTotal := metric.OperatingResultUsdHa.Mul(metric.SownAreaHa)
		rightValue = rightValue.Add(resultadoTotal)
	}

	return buildCheck(
		11,
		"Lotes",
		"Resultado operativo total",
		"Informe por cultivo",
		"∑(Resultado.Operativo_ha_lote × Superficie_lote) = ∑(Resultado.Operativo_ha_Cultivo × Superficie)",
		"∑(operating_result_per_ha × sowed_area)",
		leftValue,
		"Vista v3_lot_list",
		"∑(operating_result_usd_ha × sown_area_ha)",
		rightValue,
		"Vista v3_report_field_crop_metrics",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 12: Lotes → Informe de Resultado Generales (Resultado Operativo)
// LEFT: ∑(Resultado.Operativo_ha × Superficie) desde lotes
// RIGHT: Total Resultado Operativo del informe general
// =====================================================
func (u *UseCases) control12_LotesResultadoVsInformeGenerales(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Calcular desde lotes
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		resultadoTotal := lot.OperatingResultPerHa.Mul(lot.SowedArea)
		leftValue = leftValue.Add(resultadoTotal)
	}

	// RIGHT: Informe de Resultado Generales (primera fila = GRAL)
	filter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	if len(summaryResults) > 0 {
		rightValue = summaryResults[0].TotalOperatingResultUsd
	}

	return buildCheck(
		12,
		"Lotes",
		"Resultado operativo total",
		"Informe de Resultado Generales",
		"∑(Resultado.Operativo_ha_lote × Superficie_lote) = ∑(Resultado_Operativo_Cultivo)",
		"∑(operating_result_per_ha × sowed_area)",
		leftValue,
		"Vista v3_lot_list",
		"summaryResults[0].TotalOperatingResultUsd",
		rightValue,
		"Vista v3_report_summary_results",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 13: Lotes → Dashboard (Resultado Operativo)
// LEFT: ∑(Resultado.Operativo_ha × Superficie) desde lotes
// RIGHT: Dashboard Card Resultado Operativo
// =====================================================
func (u *UseCases) control13_LotesResultadoVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Calcular desde lotes
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, pID, 0, 0, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		resultadoTotal := lot.OperatingResultPerHa.Mul(lot.SowedArea)
		leftValue = leftValue.Add(resultadoTotal)
	}

	// RIGHT: Dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := dashboardData.Metrics.OperatingResult.ResultUSD

	return buildCheck(
		13,
		"Lotes",
		"Resultado operativo total",
		"Dashboard",
		"∑(Resultado.Operativo_ha_lote × Superficie_lote) = Resultado_Operativo Dashboard Card",
		"∑(operating_result_per_ha × sowed_area)",
		leftValue,
		"Vista v3_lot_list",
		"Metrics.OperatingResult.ResultUSD",
		rightValue,
		"Vista v3_dashboard_management_balance",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 14: Stock → Dashboard
// LEFT: Invertido - Ejecutado
// RIGHT: Dashboard.Stock
// =====================================================
func (u *UseCases) control14_StockVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	summary := dashboardData.ManagementBalance.Summary

	// LEFT: Calcular stock esperado (Invertido - Ejecutado)
	invertido := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD).
		Add(summary.LaboresInvertidosUSD)
	ejecutado := summary.DirectCostsExecutedUSD
	leftValue := invertido.Sub(ejecutado)

	// RIGHT: Stock desde dashboard
	rightValue := summary.StockUSD

	return buildCheck(
		14,
		"Stock",
		"Insumos en stock",
		"Dashboard",
		"Stock = Invertido - Ejecutado",
		"(Semillas + Agroquímicos + Fertilizantes + Labores) - Ejecutado",
		leftValue,
		"Dashboard Summary calculation",
		"Summary.StockUSD",
		rightValue,
		"Dashboard Summary",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// HELPER: buildCheck construye un IntegrityCheck con LEFT/RIGHT
// =====================================================
func buildCheck(
	controlNumber int,
	sourceModule, dataToVerify, targetModule, controlRule string,
	leftCalculation string,
	leftValue decimal.Decimal,
	leftSource string,
	rightCalculation string,
	rightValue decimal.Decimal,
	rightSource string,
	tolerance decimal.Decimal,
) domain.IntegrityCheck {
	difference := leftValue.Sub(rightValue)
	status := "OK"

	// DEBUG: Log para control 2
	if controlNumber == 2 {
		println(fmt.Sprintf("[DEBUG Control 2] leftValue: %s", leftValue.String()))
		println(fmt.Sprintf("[DEBUG Control 2] rightValue: %s", rightValue.String()))
		println(fmt.Sprintf("[DEBUG Control 2] difference: %s", difference.String()))
		println(fmt.Sprintf("[DEBUG Control 2] difference.Abs(): %s", difference.Abs().String()))
		println(fmt.Sprintf("[DEBUG Control 2] tolerance: %s", tolerance.String()))
		println(fmt.Sprintf("[DEBUG Control 2] GreaterThan result: %v", difference.Abs().GreaterThan(tolerance)))
	}

	if difference.Abs().GreaterThan(tolerance) {
		status = "ERROR"
	}

	return domain.IntegrityCheck{
		ControlNumber:    controlNumber,
		SourceModule:     sourceModule,
		DataToVerify:     dataToVerify,
		TargetModule:     targetModule,
		ControlRule:      controlRule,
		LeftCalculation:  leftCalculation,
		LeftValue:        leftValue,
		LeftSource:       leftSource,
		RightCalculation: rightCalculation,
		RightValue:       rightValue,
		RightSource:      rightSource,
		Difference:       difference,
		Status:           status,
		Tolerance:        tolerance,
	}
}
