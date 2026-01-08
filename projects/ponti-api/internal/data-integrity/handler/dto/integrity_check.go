// Package dto contiene los DTOs para el módulo de integridad de datos
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
//
// ESTOS DTOs SON CRÍTICOS PARA LA INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
package dto

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/data-integrity/usecases/domain"
)

// IntegrityReportResponse representa la respuesta del endpoint de integridad (14 controles)
type IntegrityReportResponse struct {
	Checks []IntegrityCheckDTO `json:"checks"`
}

// IntegrityCheckDTO representa un control individual de coherencia de datos
type IntegrityCheckDTO struct {
	ControlNumber int    `json:"control_number"`
	SourceModule  string `json:"source_module"`
	DataToVerify  string `json:"data_to_verify"`
	TargetModule  string `json:"target_module"`
	ControlRule   string `json:"control_rule"`

	LeftCalculation string `json:"left_calculation"`
	LeftValue       string `json:"left_value"`
	LeftSource      string `json:"left_source,omitempty"`

	RightCalculation string `json:"right_calculation"`
	RightValue       string `json:"right_value"`
	RightSource      string `json:"right_source,omitempty"`

	Difference string `json:"difference"`
	Status     string `json:"status"`
	Tolerance  string `json:"tolerance"`
}

// ToIntegrityReportResponse convierte domain.IntegrityReport a DTO
func ToIntegrityReportResponse(report *domain.IntegrityReport) *IntegrityReportResponse {
	checks := make([]IntegrityCheckDTO, len(report.Checks))

	for i, check := range report.Checks {
		checks[i] = IntegrityCheckDTO{
			ControlNumber:    check.ControlNumber,
			SourceModule:     check.SourceModule,
			DataToVerify:     check.DataToVerify,
			TargetModule:     check.TargetModule,
			ControlRule:      check.ControlRule,
			LeftCalculation:  check.LeftCalculation,
			LeftValue:        formatDecimal(check.LeftValue),
			LeftSource:       check.LeftSource,
			RightCalculation: check.RightCalculation,
			RightValue:       formatDecimal(check.RightValue),
			RightSource:      check.RightSource,
			Difference:       formatDecimal(check.Difference),
			Status:           check.Status,
			Tolerance:        formatDecimal(check.Tolerance),
		}
	}

	return &IntegrityReportResponse{
		Checks: checks,
	}
}

// formatDecimal formatea un decimal con 2 decimales
func formatDecimal(d decimal.Decimal) string {
	return d.StringFixed(2)
}
