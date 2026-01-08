package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// Nota: Decimal3 está definido en summary-results.go para evitar duplicación

// -------------------------------
// Entidades básicas de inversores
// -------------------------------

// InvestorRef: referencia mínima de un inversor (id + nombre).
type InvestorRef struct {
	InvestorID   *int64  `json:"investor_id,omitempty"`
	InvestorName *string `json:"investor_name,omitempty"`
}

// InvestorHeader: chapita de cabecera (ej: "Agrolaits 50%").
type InvestorHeader struct {
	InvestorRef
	SharePct Decimal0 `json:"share_pct"` // % global acordado - sin decimales
}

// InvestorShare: celda por inversor en una fila.
type InvestorShare struct {
	InvestorRef
	AmountUsd Decimal0 `json:"amount_usd"` // Monto USD en la celda - sin decimales
	SharePct  Decimal0 `json:"share_pct"`  // % de esa fila - sin decimales
}

// -------------------------------
// Datos generales del proyecto
// -------------------------------
type GeneralProjectData struct {
	SurfaceTotalHa Decimal0 `json:"surface_total_ha"` // Hectáreas totales - sin decimales
	LeaseFixedUsd  Decimal0 `json:"lease_fixed_usd"`  // Arriendo fijo por ha - sin decimales
	LeaseIsFixed   bool     `json:"lease_is_fixed"`   // true = arriendo fijo
	AdminPerHaUsd  Decimal0 `json:"admin_per_ha_usd"` // Administración por ha - sin decimales
	AdminTotalUsd  Decimal0 `json:"admin_total_usd"`  // Administración total - sin decimales
}

// -------------------------------
// Aportes pre-cosecha (tabla sup.)
// -------------------------------

// Enum con tipos de categoría
type ContributionCategoryType string

const (
	ContributionAgrochemicals           ContributionCategoryType = "agrochemicals"
	ContributionSeeds                   ContributionCategoryType = "seeds"
	ContributionGeneralLabors           ContributionCategoryType = "general_labors"
	ContributionSowing                  ContributionCategoryType = "sowing"
	ContributionIrrigation              ContributionCategoryType = "irrigation"
	ContributionCapitalizableLease      ContributionCategoryType = "capitalizable_lease"
	ContributionAdministrationStructure ContributionCategoryType = "administration_structure"
)

// ContributionCategory: fila de la tabla de aportes pre-cosecha
type ContributionCategory struct {
	Key                       string                   `json:"key"` // clave estable en inglés (ej: "agrochemicals")
	SortIndex                 int                      `json:"sort_index"`
	Type                      ContributionCategoryType `json:"type"`
	Label                     string                   `json:"label"`        // etiqueta visible (ej: "Agroquímicos")
	TotalUsd                  Decimal0                 `json:"total_usd"`    // Sin decimales
	TotalUsdHa                Decimal2                 `json:"total_usd_ha"` // Total u$/ha: 2 decimales
	Investors                 []InvestorShare          `json:"investors"`
	RequiresManualAttribution bool                     `json:"requires_manual_attribution"`
	AttributionNote           *string                  `json:"attribution_note,omitempty"`
}

// PreHarvestTotals: fila "Totales" en la tabla de aportes pre-cosecha
type PreHarvestTotals struct {
	TotalUsd   Decimal0        `json:"total_usd"`   // Sin decimales
	TotalUsdHa Decimal2        `json:"total_us_ha"` // Total u$/ha: 2 decimales
	Investors  []InvestorShare `json:"investors"`
}

// -------------------------------------------------
// Aporte acordado / Ajuste de aporte (bloque medio)
// -------------------------------------------------
type InvestorContributionComparison struct {
	InvestorRef
	AgreedSharePct Decimal0 `json:"agreed_share_pct"` // Sin decimales
	AgreedUsd      Decimal0 `json:"agreed_usd"`       // Sin decimales
	ActualUsd      Decimal0 `json:"actual_usd"`       // Sin decimales
	AdjustmentUsd  Decimal0 `json:"adjustment_usd"`   // Sin decimales
}

// -------------------------------
// Pagos de cosecha (tabla inferior)
// -------------------------------

type HarvestRowType string

const (
	HarvestRowHarvest HarvestRowType = "harvest" // fila detalle "Cosecha"
	HarvestRowTotals  HarvestRowType = "totals"  // fila "Totales"
)

// HarvestRow: representa una fila en pagos de cosecha
type HarvestRow struct {
	Key        string          `json:"key"`         // "harvest" o "totals"
	Type       HarvestRowType  `json:"type"`        // enum backend
	TotalUsd   Decimal0        `json:"total_usd"`   // Sin decimales
	TotalUsdHa Decimal2        `json:"total_us_ha"` // Total u$/ha: 2 decimales
	Investors  []InvestorShare `json:"investors"`
}

// HarvestSettlement: sección completa de pagos de cosecha
type HarvestSettlement struct {
	Rows                    []HarvestRow    `json:"rows"`                      // 2 filas: harvest y totals
	FooterPaymentAgreed     []InvestorShare `json:"footer_payment_agreed"`     // fila "Pago acordado"
	FooterPaymentAdjustment []InvestorShare `json:"footer_payment_adjustment"` // fila "Ajuste de pago"
}

// -------------------------------
// Informe raíz completo
// -------------------------------
type InvestorContributionReport struct {
	ProjectID       int64                            `json:"project_id"`
	ProjectName     string                           `json:"project_name"`
	CustomerID      int64                            `json:"customer_id"`
	CustomerName    string                           `json:"customer_name"`
	CampaignID      int64                            `json:"campaign_id"`
	CampaignName    string                           `json:"campaign_name"`
	InvestorHeaders []InvestorHeader                 `json:"investor_headers"`
	General         GeneralProjectData               `json:"general"`
	Contributions   []ContributionCategory           `json:"contributions"`
	PreHarvest      PreHarvestTotals                 `json:"pre_harvest"`
	Comparison      []InvestorContributionComparison `json:"comparison"`
	Harvest         HarvestSettlement                `json:"harvest"`
}

// ===== MAPPERS =====

// FromDomainInvestorReport mapea del domain al DTO
func FromDomainInvestorReport(domainReport *domain.InvestorContributionReport) *InvestorContributionReport {
	if domainReport == nil {
		return nil
	}

	return &InvestorContributionReport{
		ProjectID:       domainReport.ProjectID,
		ProjectName:     domainReport.ProjectName,
		CustomerID:      domainReport.CustomerID,
		CustomerName:    domainReport.CustomerName,
		CampaignID:      domainReport.CampaignID,
		CampaignName:    domainReport.CampaignName,
		InvestorHeaders: fromDomainInvestorHeaders(domainReport.InvestorHeaders),
		General:         fromDomainGeneralProjectData(domainReport.General),
		Contributions:   fromDomainContributionCategories(domainReport.Contributions),
		PreHarvest:      fromDomainPreHarvestTotals(domainReport.PreHarvest),
		Comparison:      fromDomainInvestorContributionComparisons(domainReport.Comparison),
		Harvest:         fromDomainHarvestSettlement(domainReport.Harvest),
	}
}

// fromDomainInvestorHeaders mapea headers de inversores
func fromDomainInvestorHeaders(domainHeaders []domain.InvestorHeader) []InvestorHeader {
	headers := make([]InvestorHeader, len(domainHeaders))
	for i, h := range domainHeaders {
		headers[i] = InvestorHeader{
			InvestorRef: InvestorRef{
				InvestorID:   h.InvestorID,
				InvestorName: h.InvestorName,
			},
			SharePct: NewDecimal0(h.SharePct), // Sin decimales
		}
	}
	return headers
}

// fromDomainGeneralProjectData mapea datos generales del proyecto
func fromDomainGeneralProjectData(domainGeneral domain.GeneralProjectData) GeneralProjectData {
	return GeneralProjectData{
		SurfaceTotalHa: NewDecimal0(domainGeneral.SurfaceTotalHa), // Sin decimales
		LeaseFixedUsd:  NewDecimal0(domainGeneral.LeaseFixedUsd),  // Sin decimales
		LeaseIsFixed:   domainGeneral.LeaseIsFixed,
		AdminPerHaUsd:  NewDecimal0(domainGeneral.AdminPerHaUsd), // Sin decimales
		AdminTotalUsd:  NewDecimal0(domainGeneral.AdminTotalUsd), // Sin decimales
	}
}

// fromDomainContributionCategories mapea categorías de contribución
func fromDomainContributionCategories(domainCategories []domain.ContributionCategory) []ContributionCategory {
	categories := make([]ContributionCategory, len(domainCategories))
	for i, c := range domainCategories {
		categories[i] = ContributionCategory{
			Key:                       c.Key,
			SortIndex:                 c.SortIndex,
			Type:                      ContributionCategoryType(c.Type),
			Label:                     c.Label,
			TotalUsd:                  NewDecimal0(c.TotalUsd),   // Sin decimales
			TotalUsdHa:                NewDecimal2(c.TotalUsdHa), // Total u$/ha: 2 decimales
			Investors:                 fromDomainInvestorShares(c.Investors),
			RequiresManualAttribution: c.RequiresManualAttribution,
			AttributionNote:           c.AttributionNote,
		}
	}
	return categories
}

// fromDomainPreHarvestTotals mapea totales pre-cosecha
func fromDomainPreHarvestTotals(domainTotals domain.PreHarvestTotals) PreHarvestTotals {
	return PreHarvestTotals{
		TotalUsd:   NewDecimal0(domainTotals.TotalUsd),   // Sin decimales
		TotalUsdHa: NewDecimal2(domainTotals.TotalUsdHa), // Total u$/ha: 2 decimales
		Investors:  fromDomainInvestorShares(domainTotals.Investors),
	}
}

// fromDomainInvestorContributionComparisons mapea comparaciones de contribución
func fromDomainInvestorContributionComparisons(domainComparisons []domain.InvestorContributionComparison) []InvestorContributionComparison {
	comparisons := make([]InvestorContributionComparison, len(domainComparisons))
	for i, c := range domainComparisons {
		comparisons[i] = InvestorContributionComparison{
			InvestorRef: InvestorRef{
				InvestorID:   c.InvestorID,
				InvestorName: c.InvestorName,
			},
			AgreedSharePct: NewDecimal0(c.AgreedSharePct), // Sin decimales
			AgreedUsd:      NewDecimal0(c.AgreedUsd),      // Sin decimales
			ActualUsd:      NewDecimal0(c.ActualUsd),      // Sin decimales
			AdjustmentUsd:  NewDecimal0(c.AdjustmentUsd),  // Sin decimales
		}
	}
	return comparisons
}

// fromDomainHarvestSettlement mapea liquidación de cosecha
func fromDomainHarvestSettlement(domainHarvest domain.HarvestSettlement) HarvestSettlement {
	return HarvestSettlement{
		Rows:                    fromDomainHarvestRows(domainHarvest.Rows),
		FooterPaymentAgreed:     fromDomainInvestorShares(domainHarvest.FooterPaymentAgreed),
		FooterPaymentAdjustment: fromDomainInvestorShares(domainHarvest.FooterPaymentAdjustment),
	}
}

// fromDomainHarvestRows mapea filas de cosecha
func fromDomainHarvestRows(domainRows []domain.HarvestRow) []HarvestRow {
	rows := make([]HarvestRow, len(domainRows))
	for i, r := range domainRows {
		rows[i] = HarvestRow{
			Key:        r.Key,
			Type:       HarvestRowType(r.Type),
			TotalUsd:   NewDecimal0(r.TotalUsd),   // Sin decimales
			TotalUsdHa: NewDecimal2(r.TotalUsdHa), // Total u$/ha: 2 decimales
			Investors:  fromDomainInvestorShares(r.Investors),
		}
	}
	return rows
}

// fromDomainInvestorShares mapea shares de inversores
func fromDomainInvestorShares(domainShares []domain.InvestorShare) []InvestorShare {
	shares := make([]InvestorShare, len(domainShares))
	for i, s := range domainShares {
		shares[i] = InvestorShare{
			InvestorRef: InvestorRef{
				InvestorID:   s.InvestorID,
				InvestorName: s.InvestorName,
			},
			AmountUsd: NewDecimal0(s.AmountUsd), // Sin decimales
			SharePct:  NewDecimal0(s.SharePct),  // Sin decimales
		}
	}
	return shares
}
