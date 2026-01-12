// Package db provee helpers para acceso a base de datos
package db

import "os"

// reportSchema contiene el schema a usar para vistas de reportes
// Si REPORT_SCHEMA="v4_report" → usa vistas v4_report.*
// Si REPORT_SCHEMA está vacío o es "public" → usa vistas v3_* (default)
var reportSchema = os.Getenv("REPORT_SCHEMA")

// ReportView retorna el nombre completo de la vista según el schema configurado
// Ejemplo: ReportView("lot_metrics") → "v4_report.lot_metrics" o "v3_lot_metrics"
func ReportView(name string) string {
	if reportSchema == "v4_report" {
		return "v4_report." + name
	}
	return "v3_" + name
}

// FieldCropView retorna el nombre de vista para field_crop
// v3: v3_report_field_crop_{name} → v4: v4_report.field_crop_{name}
// Ejemplo: FieldCropView("metrics") → "v4_report.field_crop_metrics" o "v3_report_field_crop_metrics"
func FieldCropView(name string) string {
	if reportSchema == "v4_report" {
		return "v4_report.field_crop_" + name
	}
	return "v3_report_field_crop_" + name
}

// SummaryView retorna el nombre de vista para summary_results
// v3: v3_report_summary_results_view → v4: v4_report.summary_results
func SummaryView() string {
	if reportSchema == "v4_report" {
		return "v4_report.summary_results"
	}
	return "v3_report_summary_results_view"
}

// IsV4Enabled retorna true si está habilitado el schema v4
func IsV4Enabled() bool {
	return reportSchema == "v4_report"
}

// DashboardView retorna el nombre de vista para dashboard
// v3: v3_dashboard_{name} → v4: v4_report.dashboard_{name}
// Ejemplo: DashboardView("metrics") → "v4_report.dashboard_metrics" o "v3_dashboard_metrics"
func DashboardView(name string) string {
	if reportSchema == "v4_report" {
		return "v4_report.dashboard_" + name
	}
	return "v3_dashboard_" + name
}

// InvestorView retorna el nombre de vista para investor
// v3: v3_investor_{name} o v3_report_investor_{name} → v4: v4_report.investor_{name}
// Ejemplo: InvestorView("contribution_data") → "v4_report.investor_contribution_data" o "v3_investor_contribution_data_view"
func InvestorView(name string) string {
	if reportSchema == "v4_report" {
		return "v4_report.investor_" + name
	}
	// Mapeo especial para v3 porque tienen nombres diferentes
	switch name {
	case "contribution_data":
		return "v3_investor_contribution_data_view"
	case "project_base":
		return "v3_report_investor_project_base"
	case "contribution_categories":
		return "v3_report_investor_contribution_categories"
	case "distributions":
		return "v3_report_investor_distributions"
	default:
		return "v3_investor_" + name
	}
}
