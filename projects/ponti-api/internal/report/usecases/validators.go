// Package usecases proporciona validadores para los casos de uso de reportes
package usecases

import (
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// ReportFilterValidator valida filtros de reportes
type ReportFilterValidator struct{}

// NewReportFilterValidator crea una nueva instancia del validador
func NewReportFilterValidator() *ReportFilterValidator {
	return &ReportFilterValidator{}
}

// ValidateAtLeastOneFilter valida que al menos un filtro esté presente
func (v *ReportFilterValidator) ValidateAtLeastOneFilter(filter domain.SummaryResultsFilter) error {
	if filter.ProjectID == nil &&
		filter.CustomerID == nil &&
		filter.CampaignID == nil &&
		filter.FieldID == nil {
		return fmt.Errorf("at least one filter must be specified (project_id, customer_id, campaign_id, or field_id)")
	}
	return nil
}

// HasAnyFilter verifica si hay al menos un filtro presente
func (v *ReportFilterValidator) HasAnyFilter(filter domain.SummaryResultsFilter) bool {
	return filter.ProjectID != nil ||
		filter.CustomerID != nil ||
		filter.CampaignID != nil ||
		filter.FieldID != nil
}

