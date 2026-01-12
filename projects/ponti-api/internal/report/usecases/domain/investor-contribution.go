// Package domain: modelos de dominio para el informe de aportes de inversores.
package domain

import "github.com/shopspring/decimal"

// =====================================================================================
// ENUMS / CONSTANTES
// =====================================================================================

// ContributionCategoryType tipos de filas de la tabla "Aportes pre-cosecha"
// Deben mapear 1:1 con la maqueta (Agroquímicos, Semilla, Labores grales, etc.).
type ContributionCategoryType string

const (
	ContributionAgrochemicals           ContributionCategoryType = "agrochemicals"            // Agroquímicos
	ContributionSeeds                   ContributionCategoryType = "seeds"                    // Semilla
	ContributionGeneralLabors           ContributionCategoryType = "general_labors"           // Labores grales
	ContributionSowing                  ContributionCategoryType = "sowing"                   // Siembra
	ContributionIrrigation              ContributionCategoryType = "irrigation"               // Riego
	ContributionCapitalizableLease      ContributionCategoryType = "capitalizable_lease"      // Arriendo (capitalizable/fijo según el proyecto)
	ContributionAdministrationStructure ContributionCategoryType = "administration_structure" // Administración y estructura
)

// =====================================================================================
// OBJETOS BÁSICOS (INVERSORES)
// =====================================================================================

// InvestorRef referencia mínima a un inversor (ID/nombre).
// Se usa en varias estructuras para no repetir campos.
type InvestorRef struct {
	InvestorID   *int64
	InvestorName *string
}

// InvestorHeader datos para la "chapita" de cabecera por inversor (p. ej. "Agrolaits 50%").
// La UI usa esta lista para mostrar las cabeceras sobre las columnas de inversores.
type InvestorHeader struct {
	InvestorRef
	SharePct decimal.Decimal // Porcentaje global acordado (0..100) que se muestra en la cabecera.
}

// InvestorShare asignación monetaria por inversor dentro de una fila/categoría.
// "share_pct" es el % de esa FILA (no el global de cabecera).
type InvestorShare struct {
	InvestorRef
	AmountUsd decimal.Decimal // Monto USD asignado al inversor en esa fila.
	SharePct  decimal.Decimal // % dentro de la categoría/fila (0..100).
}

// =====================================================================================
// DATOS GENERALES DEL PROYECTO (encabezado del informe)
// =====================================================================================

// GeneralProjectData datos base del proyecto que afectan el informe.
// En la maqueta se usan para costos por ha (admin, arriendo fijo, etc.).
type GeneralProjectData struct {
	SurfaceTotalHa decimal.Decimal // Superficie total del proyecto (ha).
	LeaseFixedUsd  decimal.Decimal // Arriendo fijo por ha (si aplica en el proyecto).
	LeaseIsFixed   bool            // true => arriendo fijo (no prorratea con % de inversores).
	AdminPerHaUsd  decimal.Decimal // Costo de administración por ha.
	AdminTotalUsd  decimal.Decimal // Costo de administración total.
}

// =====================================================================================
// APORTES PRE-COSECHA (tabla principal)
// =====================================================================================

// ContributionCategory representa UNA fila de la tabla "Aportes pre-cosecha".
// - sort_index mantiene el orden visual igual a la maqueta.
// - investors contiene las columnas por inversor con monto y % de esa fila.
// - total_usd y total_usd_ha corresponden a las columnas "TOTAL US" y "TOTAL US/HA".
type ContributionCategory struct {
	Key                       string                   // clave estable en inglés (ej: "agrochemicals")
	SortIndex                 int                      // Orden de la fila en la tabla.
	Type                      ContributionCategoryType // Tipo de categoría (enum).
	Label                     string                   // Texto a mostrar (p. ej., "Agroquímicos").
	TotalUsd                  decimal.Decimal          // Total USD de la fila.
	TotalUsdHa                decimal.Decimal          // Total USD/ha de la fila.
	Investors                 []InvestorShare          // Columnas por inversor en esta fila.
	RequiresManualAttribution bool                     // true si la fila requiere asignación manual (100/0, etc.).
	AttributionNote           *string                  // Texto opcional para explicar la regla de asignación.
}

// PreHarvestTotals corresponde a la FILA "Totales" de la sección "Aportes pre-cosecha".
// Incluye totales generales y el desglose por inversor de esa fila.
type PreHarvestTotals struct {
	TotalUsd   decimal.Decimal // Suma de todas las filas (columna TOTAL US).
	TotalUsdHa decimal.Decimal // Suma por ha (columna TOTAL US/HA).
	Investors  []InvestorShare // Totales por inversor (columnas de inversores en la fila "Totales").
}

// =====================================================================================
// APORTE ACORDADO / AJUSTE DE APORTE (bloque bajo la tabla)
// =====================================================================================

// InvestorContributionComparison compara lo acordado vs lo efectivamente aportado.
// Se usa para renderizar "Aporte acordado" y "Ajuste de aporte" por inversor.
type InvestorContributionComparison struct {
	InvestorRef
	AgreedSharePct decimal.Decimal // % acordado global (0..100) para el inversor.
	AgreedUsd      decimal.Decimal // Monto acordado total (USD).
	ActualUsd      decimal.Decimal // Monto efectivamente aportado (USD).
	AdjustmentUsd  decimal.Decimal // Diferencia: AgreedUsd - ActualUsd (positivo => debe aportar).
}

// =====================================================================================
// PAGOS DE COSECHA (sección inferior)
// =====================================================================================

type HarvestRowType string

const (
	HarvestRowHarvest HarvestRowType = "harvest" // fila detalle "Cosecha"
	HarvestRowTotals  HarvestRowType = "totals"  // fila "Totales"
)

// HarvestRow representa una fila en pagos de cosecha
type HarvestRow struct {
	Key        string         // "harvest" o "totals"
	Type       HarvestRowType // enum backend
	TotalUsd   decimal.Decimal
	TotalUsdHa decimal.Decimal
	Investors  []InvestorShare
}

// HarvestSettlement sección completa de pagos de cosecha
type HarvestSettlement struct {
	Rows                    []HarvestRow    // 2 filas: harvest y totals
	FooterPaymentAgreed     []InvestorShare // fila "Pago acordado"
	FooterPaymentAdjustment []InvestorShare // fila "Ajuste de pago"
}

// =====================================================================================
// INFORME COMPLETO (root DTO consumido por la UI)
// =====================================================================================

// InvestorContributionReport objeto raíz que consolida todas las secciones del informe.
type InvestorContributionReport struct {
	// ---- Identificación / metadatos del proyecto ----
	ProjectID    int64  // ID interno del proyecto.
	ProjectName  string // Nombre del proyecto.
	CustomerID   int64  // ID del cliente/propietario.
	CustomerName string // Nombre del cliente.
	CampaignID   int64  // ID de campaña.
	CampaignName string // Nombre de campaña.

	// ---- Cabecera de inversores (chapitas con % global) ----
	InvestorHeaders []InvestorHeader // Ordenadas de izquierda a derecha como en la maqueta.

	// ---- Datos generales del proyecto ----
	General GeneralProjectData // Costos por ha y totales relevantes.

	// ---- Aportes pre-cosecha ----
	Contributions []ContributionCategory // Filas de la tabla (en el orden provisto por sort_index).
	PreHarvest    PreHarvestTotals       // Fila "Totales" de la tabla.

	// ---- Aporte acordado / Ajuste de aporte ----
	Comparison []InvestorContributionComparison // Una entrada por inversor.

	// ---- Pagos de cosecha ----
	Harvest HarvestSettlement // Totales de cosecha + liquidación por inversor.
}
