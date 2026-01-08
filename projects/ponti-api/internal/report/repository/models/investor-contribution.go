package models

import (
	"encoding/json"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== MODELOS PARA APORTES DE INVERSORES =====

// InvestorContributionDataModel modelo para los datos de aportes de inversores desde la vista
// La vista v3_investor_contribution_data_view devuelve datos en formato JSONB
type InvestorContributionDataModel struct {
	ProjectID    int64  `gorm:"column:project_id"`
	ProjectName  string `gorm:"column:project_name"`
	CustomerID   int64  `gorm:"column:customer_id"`
	CustomerName string `gorm:"column:customer_name"`
	CampaignID   int64  `gorm:"column:campaign_id"`
	CampaignName string `gorm:"column:campaign_name"`

	// Datos en formato JSONB desde la vista
	InvestorHeadersJSON                string `gorm:"column:investor_headers"`
	GeneralProjectDataJSON             string `gorm:"column:general_project_data"`
	ContributionCategoriesJSON         string `gorm:"column:contribution_categories"`
	InvestorContributionComparisonJSON string `gorm:"column:investor_contribution_comparison"`
	HarvestSettlementJSON              string `gorm:"column:harvest_settlement"`
}

// InvestorHeaderModel modelo para headers de inversores
type InvestorHeaderModel struct {
	InvestorID   *int64          `json:"investor_id,omitempty"`
	InvestorName *string         `json:"investor_name,omitempty"`
	SharePct     decimal.Decimal `json:"share_pct"`
}

// InvestorShareModel modelo para shares de inversores
type InvestorShareModel struct {
	InvestorID   *int64          `json:"investor_id,omitempty"`
	InvestorName *string         `json:"investor_name,omitempty"`
	AmountUsd    decimal.Decimal `json:"amount_usd"`
	SharePct     decimal.Decimal `json:"share_pct"`
}

// GeneralProjectDataModel modelo para datos generales del proyecto
type GeneralProjectDataModel struct {
	SurfaceTotalHa decimal.Decimal `json:"surface_total_ha"`
	LeaseFixedUsd  decimal.Decimal `json:"lease_fixed_total_usd"` // Corregido: coincide con vista SQL
	LeaseIsFixed   bool            `json:"lease_is_fixed"`
	AdminPerHaUsd  decimal.Decimal `json:"admin_per_ha_usd"`
	AdminTotalUsd  decimal.Decimal `json:"admin_total_usd"`
}

// ContributionCategoryModel modelo para categorías de contribución
type ContributionCategoryModel struct {
	Key                       string               `json:"key"`
	SortIndex                 int                  `json:"sort_index"`
	Type                      string               `json:"type"`
	Label                     string               `json:"label"`
	TotalUsd                  decimal.Decimal      `json:"total_usd"`
	TotalUsdHa                decimal.Decimal      `json:"total_usd_ha"`
	Investors                 []InvestorShareModel `json:"investors"`
	RequiresManualAttribution bool                 `json:"requires_manual_attribution"`
	AttributionNote           *string              `json:"attribution_note,omitempty"`
}

// PreHarvestTotalsModel modelo para totales pre-cosecha
type PreHarvestTotalsModel struct {
	TotalUsd   decimal.Decimal      `json:"total_usd"`
	TotalUsdHa decimal.Decimal      `json:"total_us_ha"`
	Investors  []InvestorShareModel `json:"investors"`
}

// InvestorContributionComparisonModel modelo para comparaciones de contribución
type InvestorContributionComparisonModel struct {
	InvestorID     *int64          `json:"investor_id,omitempty"`
	InvestorName   *string         `json:"investor_name,omitempty"`
	AgreedSharePct decimal.Decimal `json:"agreed_share_pct"` // FIX 000172: Corregido para coincidir con vista SQL
	AgreedUsd      decimal.Decimal `json:"agreed_usd"`       // FIX 000172: Corregido para coincidir con vista SQL
	ActualUsd      decimal.Decimal `json:"actual_usd"`       // FIX 000172: Corregido para coincidir con vista SQL
	AdjustmentUsd  decimal.Decimal `json:"adjustment_usd"`
}

// HarvestRowModel modelo para filas de cosecha
type HarvestRowModel struct {
	Key        string               `json:"key"`
	Type       string               `json:"type"`
	TotalUsd   decimal.Decimal      `json:"total_usd"`
	TotalUsdHa decimal.Decimal      `json:"total_us_ha"`
	Investors  []InvestorShareModel `json:"investors"`
}

// HarvestSettlementModel modelo para liquidación de cosecha
type HarvestSettlementModel struct {
	Rows                    []HarvestRowModel    `json:"rows"`
	FooterPaymentAgreed     []InvestorShareModel `json:"footer_payment_agreed"`
	FooterPaymentAdjustment []InvestorShareModel `json:"footer_payment_adjustment"`
}

// InvestorContributionReportModel modelo completo del reporte
type InvestorContributionReportModel struct {
	ProjectID       int64                                 `json:"project_id"`
	ProjectName     string                                `json:"project_name"`
	CustomerID      int64                                 `json:"customer_id"`
	CustomerName    string                                `json:"customer_name"`
	CampaignID      int64                                 `json:"campaign_id"`
	CampaignName    string                                `json:"campaign_name"`
	InvestorHeaders []InvestorHeaderModel                 `json:"investor_headers"`
	General         GeneralProjectDataModel               `json:"general"`
	Contributions   []ContributionCategoryModel           `json:"contributions"`
	PreHarvest      PreHarvestTotalsModel                 `json:"pre_harvest"`
	Comparison      []InvestorContributionComparisonModel `json:"comparison"`
	Harvest         HarvestSettlementModel                `json:"harvest"`
}

// ===== MAPPERS =====

// ToDomainInvestorContributionReport convierte el modelo al domain
// Deserializa los datos JSONB de la vista v3_investor_contribution_data_view
func (m *InvestorContributionDataModel) ToDomainInvestorContributionReport() (*domain.InvestorContributionReport, error) {
	report := &domain.InvestorContributionReport{
		ProjectID:    m.ProjectID,
		ProjectName:  m.ProjectName,
		CustomerID:   m.CustomerID,
		CustomerName: m.CustomerName,
		CampaignID:   m.CampaignID,
		CampaignName: m.CampaignName,
	}

	// Parsear investor headers desde JSONB
	if m.InvestorHeadersJSON != "" && m.InvestorHeadersJSON != "null" {
		var headers []InvestorHeaderModel
		if err := json.Unmarshal([]byte(m.InvestorHeadersJSON), &headers); err != nil {
			return nil, fmt.Errorf("error deserializando investor_headers: %w (JSON: %s)", err, m.InvestorHeadersJSON)
		}
		report.InvestorHeaders = m.mapInvestorHeadersToDomain(headers)
	}

	// Parsear datos generales del proyecto desde JSONB
	if m.GeneralProjectDataJSON != "" {
		var generalData GeneralProjectDataModel
		if err := json.Unmarshal([]byte(m.GeneralProjectDataJSON), &generalData); err != nil {
			return nil, fmt.Errorf("error deserializando general_project_data: %w", err)
		}
		report.General = domain.GeneralProjectData{
			SurfaceTotalHa: generalData.SurfaceTotalHa,
			LeaseFixedUsd:  generalData.LeaseFixedUsd,
			LeaseIsFixed:   generalData.LeaseIsFixed,
			AdminPerHaUsd:  generalData.AdminPerHaUsd,
			AdminTotalUsd:  generalData.AdminTotalUsd,
		}
	}

	// Parsear contributions desde JSONB
	if m.ContributionCategoriesJSON != "" {
		var contributions []ContributionCategoryModel
		if err := json.Unmarshal([]byte(m.ContributionCategoriesJSON), &contributions); err != nil {
			return nil, fmt.Errorf("error deserializando contribution_categories: %w", err)
		}
		report.Contributions = m.mapContributionsToDomain(contributions)

		// Calcular PreHarvest sumando todas las categorías
		report.PreHarvest = m.calculatePreHarvestFromContributions(contributions)
	}

	// Parsear comparison desde JSONB
	if m.InvestorContributionComparisonJSON != "" {
		var comparisons []InvestorContributionComparisonModel
		if err := json.Unmarshal([]byte(m.InvestorContributionComparisonJSON), &comparisons); err != nil {
			return nil, fmt.Errorf("error deserializando investor_contribution_comparison: %w", err)
		}
		report.Comparison = m.mapComparisonsToDomain(comparisons)
	}

	// Parsear harvest desde JSONB
	if m.HarvestSettlementJSON != "" {
		var harvest HarvestSettlementModel
		if err := json.Unmarshal([]byte(m.HarvestSettlementJSON), &harvest); err != nil {
			return nil, fmt.Errorf("error deserializando harvest_settlement: %w", err)
		}
		report.Harvest = m.mapHarvestToDomain(harvest)
	}

	return report, nil
}

// mapInvestorHeadersToDomain mapea headers de inversores del modelo al domain
func (m *InvestorContributionDataModel) mapInvestorHeadersToDomain(headers []InvestorHeaderModel) []domain.InvestorHeader {
	domainHeaders := make([]domain.InvestorHeader, len(headers))
	for i, h := range headers {
		domainHeaders[i] = domain.InvestorHeader{
			InvestorRef: domain.InvestorRef{
				InvestorID:   h.InvestorID,
				InvestorName: h.InvestorName,
			},
			SharePct: h.SharePct,
		}
	}
	return domainHeaders
}

// mapContributionsToDomain mapea contribuciones del modelo al domain usando helpers (DRY)
func (m *InvestorContributionDataModel) mapContributionsToDomain(contributions []ContributionCategoryModel) []domain.ContributionCategory {
	domainContributions := make([]domain.ContributionCategory, len(contributions))
	for i, c := range contributions {
		domainContributions[i] = domain.ContributionCategory{
			Key:                       c.Key,
			SortIndex:                 c.SortIndex,
			Type:                      domain.ContributionCategoryType(c.Type),
			Label:                     c.Label,
			TotalUsd:                  c.TotalUsd,
			TotalUsdHa:                c.TotalUsdHa,
			Investors:                 ConvertInvestorSharesSlice(c.Investors),
			RequiresManualAttribution: c.RequiresManualAttribution,
			AttributionNote:           c.AttributionNote,
		}
	}
	return domainContributions
}

// mapComparisonsToDomain mapea comparaciones del modelo al domain
func (m *InvestorContributionDataModel) mapComparisonsToDomain(comparisons []InvestorContributionComparisonModel) []domain.InvestorContributionComparison {
	domainComparisons := make([]domain.InvestorContributionComparison, len(comparisons))
	for i, c := range comparisons {
		domainComparisons[i] = domain.InvestorContributionComparison{
			InvestorRef: domain.InvestorRef{
				InvestorID:   c.InvestorID,
				InvestorName: c.InvestorName,
			},
			AgreedSharePct: c.AgreedSharePct,
			AgreedUsd:      c.AgreedUsd,
			ActualUsd:      c.ActualUsd,
			AdjustmentUsd:  c.AdjustmentUsd,
		}
	}
	return domainComparisons
}

// mapHarvestToDomain mapea harvest del modelo al domain usando helpers (DRY)
func convertInvestorSharesWithTotal(shares []InvestorShareModel, total decimal.Decimal) []domain.InvestorShare {
	result := make([]domain.InvestorShare, len(shares))

	hasTotal := !total.IsZero()

	for i, share := range shares {
		percentage := share.SharePct
		if hasTotal {
			percentage = share.AmountUsd.Div(total).Mul(decimal.NewFromInt(100))
		}

		result[i] = domain.InvestorShare{
			InvestorRef: domain.InvestorRef{
				InvestorID:   share.InvestorID,
				InvestorName: share.InvestorName,
			},
			AmountUsd: share.AmountUsd,
			SharePct:  percentage,
		}
	}

	return result
}

func (m *InvestorContributionDataModel) mapHarvestToDomain(harvest HarvestSettlementModel) domain.HarvestSettlement {
	domainRows := make([]domain.HarvestRow, len(harvest.Rows))
	for i, r := range harvest.Rows {
		domainRows[i] = domain.HarvestRow{
			Key:        r.Key,
			Type:       domain.HarvestRowType(r.Type),
			TotalUsd:   r.TotalUsd,
			TotalUsdHa: r.TotalUsdHa,
			Investors:  convertInvestorSharesWithTotal(r.Investors, r.TotalUsd),
		}
	}

	return domain.HarvestSettlement{
		Rows:                    domainRows,
		FooterPaymentAgreed:     ConvertInvestorSharesSlice(harvest.FooterPaymentAgreed),
		FooterPaymentAdjustment: ConvertInvestorSharesSlice(harvest.FooterPaymentAdjustment),
	}
}

// calculatePreHarvestFromContributions calcula los totales de pre-harvest sumando todas las categorías
func (m *InvestorContributionDataModel) calculatePreHarvestFromContributions(contributions []ContributionCategoryModel) domain.PreHarvestTotals {
	var totalUsd decimal.Decimal
	var totalUsdHa decimal.Decimal

	// Headers determinan el orden de columnas en la UI
	var headerOrder []InvestorHeaderModel
	if m.InvestorHeadersJSON != "" && m.InvestorHeadersJSON != "null" {
		if err := json.Unmarshal([]byte(m.InvestorHeadersJSON), &headerOrder); err == nil {
			for _, header := range headerOrder {
				if header.InvestorID == nil {
					continue
				}
			}
		}
	}

	// Mapa para acumular montos por inversor
	investorTotals := make(map[int64]*domain.InvestorShare)

	for _, category := range contributions {
		totalUsd = totalUsd.Add(category.TotalUsd)
		totalUsdHa = totalUsdHa.Add(category.TotalUsdHa)

		// Sumar los montos de cada inversor
		for _, investor := range category.Investors {
			if investor.InvestorID == nil {
				continue
			}

			if _, exists := investorTotals[*investor.InvestorID]; !exists {
				investorTotals[*investor.InvestorID] = &domain.InvestorShare{
					InvestorRef: domain.InvestorRef{
						InvestorID:   investor.InvestorID,
						InvestorName: investor.InvestorName,
					},
					AmountUsd: decimal.Zero,
					SharePct:  decimal.Zero,
				}
			}
			investorTotals[*investor.InvestorID].AmountUsd = investorTotals[*investor.InvestorID].AmountUsd.Add(investor.AmountUsd)
		}
	}

	// Convertir el mapa a slice y calcular porcentajes
	investors := make([]domain.InvestorShare, 0, len(investorTotals))

	calcSharePct := func(amount decimal.Decimal) decimal.Decimal {
		if totalUsd.IsZero() {
			return decimal.Zero
		}
		return amount.Div(totalUsd).Mul(decimal.NewFromInt(100))
	}

	if len(headerOrder) > 0 {
		// Respetar el orden de los headers (clientes y sociedades)
		used := make(map[int64]struct{})
		for _, header := range headerOrder {
			if header.InvestorID == nil {
				continue
			}

			var amount decimal.Decimal
			if inv, ok := investorTotals[*header.InvestorID]; ok {
				amount = inv.AmountUsd
				used[*header.InvestorID] = struct{}{}
			}

			investors = append(investors, domain.InvestorShare{
				InvestorRef: domain.InvestorRef{
					InvestorID:   header.InvestorID,
					InvestorName: header.InvestorName,
				},
				AmountUsd: amount,
				SharePct:  calcSharePct(amount),
			})
		}

		// Agregar cualquier inversor adicional que no estuviera en los headers
		for id, inv := range investorTotals {
			if _, ok := used[id]; ok {
				continue
			}

			inv.SharePct = calcSharePct(inv.AmountUsd)
			investors = append(investors, *inv)
		}
	} else {
		for _, inv := range investorTotals {
			inv.SharePct = calcSharePct(inv.AmountUsd)
			investors = append(investors, *inv)
		}
	}

	return domain.PreHarvestTotals{
		TotalUsd:   totalUsd,
		TotalUsdHa: totalUsdHa,
		Investors:  investors,
	}
}
