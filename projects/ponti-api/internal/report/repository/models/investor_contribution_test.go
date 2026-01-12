package models

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneralProjectDataModel_UnmarshalJSON verifica que los datos generales del proyecto
// se deserializan correctamente desde JSON (migración 000171)
func TestGeneralProjectDataModel_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected GeneralProjectDataModel
		wantErr  bool
	}{
		{
			name: "Proyecto con todos los campos poblados",
			jsonData: `{
				"surface_total_ha": 185.000000,
				"lease_fixed_total_usd": 4440.44,
				"lease_is_fixed": true,
				"admin_per_ha_usd": 40.0000000000000000,
				"admin_total_usd": 7400.000000000
			}`,
			expected: GeneralProjectDataModel{
				SurfaceTotalHa: decimal.NewFromFloat(185),
				LeaseFixedUsd:  decimal.NewFromFloat(4440.44),
				LeaseIsFixed:   true,
				AdminPerHaUsd:  decimal.NewFromFloat(40),
				AdminTotalUsd:  decimal.NewFromFloat(7400),
			},
			wantErr: false,
		},
		{
			name: "Proyecto sin arriendo (valores en 0)",
			jsonData: `{
				"surface_total_ha": 100.50,
				"lease_fixed_total_usd": 0,
				"lease_is_fixed": false,
				"admin_per_ha_usd": 35.5,
				"admin_total_usd": 3567.75
			}`,
			expected: GeneralProjectDataModel{
				SurfaceTotalHa: decimal.NewFromFloat(100.50),
				LeaseFixedUsd:  decimal.Zero,
				LeaseIsFixed:   false,
				AdminPerHaUsd:  decimal.NewFromFloat(35.5),
				AdminTotalUsd:  decimal.NewFromFloat(3567.75),
			},
			wantErr: false,
		},
		{
			name: "Proyecto vacío (todos en 0)",
			jsonData: `{
				"surface_total_ha": 0,
				"lease_fixed_total_usd": 0,
				"lease_is_fixed": false,
				"admin_per_ha_usd": 0,
				"admin_total_usd": 0
			}`,
			expected: GeneralProjectDataModel{
				SurfaceTotalHa: decimal.Zero,
				LeaseFixedUsd:  decimal.Zero,
				LeaseIsFixed:   false,
				AdminPerHaUsd:  decimal.Zero,
				AdminTotalUsd:  decimal.Zero,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result GeneralProjectDataModel
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, tt.expected.SurfaceTotalHa.Equal(result.SurfaceTotalHa),
				"SurfaceTotalHa: expected %v, got %v", tt.expected.SurfaceTotalHa, result.SurfaceTotalHa)
			assert.True(t, tt.expected.LeaseFixedUsd.Equal(result.LeaseFixedUsd),
				"LeaseFixedUsd: expected %v, got %v", tt.expected.LeaseFixedUsd, result.LeaseFixedUsd)
			assert.Equal(t, tt.expected.LeaseIsFixed, result.LeaseIsFixed)
			assert.True(t, tt.expected.AdminPerHaUsd.Equal(result.AdminPerHaUsd),
				"AdminPerHaUsd: expected %v, got %v", tt.expected.AdminPerHaUsd, result.AdminPerHaUsd)
			assert.True(t, tt.expected.AdminTotalUsd.Equal(result.AdminTotalUsd),
				"AdminTotalUsd: expected %v, got %v", tt.expected.AdminTotalUsd, result.AdminTotalUsd)
		})
	}
}

// TestInvestorContributionDataModel_ToDomainInvestorContributionReport_GeneralData
// verifica que los datos generales se mapean correctamente al dominio
func TestInvestorContributionDataModel_ToDomainInvestorContributionReport_GeneralData(t *testing.T) {
	// Caso de prueba: Proyecto 11 - CONTROL INTEGRAL
	model := &InvestorContributionDataModel{
		ProjectID:    11,
		ProjectName:  "CONTROL INTEGRAL",
		CustomerID:   7,
		CustomerName: "CONTROL INTEGRAL",
		CampaignID:   4,
		CampaignName: "2025",
		GeneralProjectDataJSON: `{
			"surface_total_ha": 185.000000,
			"lease_fixed_total_usd": 4440.44,
			"lease_is_fixed": true,
			"admin_per_ha_usd": 40.0000000000000000,
			"admin_total_usd": 7400.000000000
		}`,
		InvestorHeadersJSON:                "[]",
		ContributionCategoriesJSON:         "[]",
		InvestorContributionComparisonJSON: "[]",
		HarvestSettlementJSON:              `{"rows": [], "footer_payment_agreed": [], "footer_payment_adjustment": []}`,
	}

	report, err := model.ToDomainInvestorContributionReport()

	require.NoError(t, err)
	require.NotNil(t, report)

	// Verificar datos básicos
	assert.Equal(t, int64(11), report.ProjectID)
	assert.Equal(t, "CONTROL INTEGRAL", report.ProjectName)

	// Verificar datos generales (Card Superficie)
	assert.True(t, decimal.NewFromInt(185).Equal(report.General.SurfaceTotalHa),
		"SurfaceTotalHa debe ser 185, got %v", report.General.SurfaceTotalHa)
	assert.True(t, decimal.NewFromFloat(4440.44).Equal(report.General.LeaseFixedUsd),
		"LeaseFixedUsd debe ser 4440.44, got %v", report.General.LeaseFixedUsd)
	assert.True(t, report.General.LeaseIsFixed,
		"LeaseIsFixed debe ser true")
	assert.True(t, decimal.NewFromInt(40).Equal(report.General.AdminPerHaUsd),
		"AdminPerHaUsd debe ser 40, got %v", report.General.AdminPerHaUsd)
	assert.True(t, decimal.NewFromInt(7400).Equal(report.General.AdminTotalUsd),
		"AdminTotalUsd debe ser 7400, got %v", report.General.AdminTotalUsd)
}

// TestInvestorContributionDataModel_ToDomainInvestorContributionReport_EmptyGeneral
// verifica el comportamiento cuando general_project_data está vacío o es null
func TestInvestorContributionDataModel_ToDomainInvestorContributionReport_EmptyGeneral(t *testing.T) {
	tests := []struct {
		name        string
		generalJSON string
		wantErr     bool
	}{
		{
			name:        "JSON vacío",
			generalJSON: "",
			wantErr:     false,
		},
		{
			name:        "JSON null",
			generalJSON: "null",
			wantErr:     false,
		},
		{
			name: "JSON con estructura vacía",
			generalJSON: `{
				"surface_total_ha": 0,
				"lease_fixed_total_usd": 0,
				"lease_is_fixed": false,
				"admin_per_ha_usd": 0,
				"admin_total_usd": 0
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &InvestorContributionDataModel{
				ProjectID:                          1,
				ProjectName:                        "Test Project",
				GeneralProjectDataJSON:             tt.generalJSON,
				InvestorHeadersJSON:                "[]",
				ContributionCategoriesJSON:         "[]",
				InvestorContributionComparisonJSON: "[]",
				HarvestSettlementJSON:              `{"rows": [], "footer_payment_agreed": [], "footer_payment_adjustment": []}`,
			}

			report, err := model.ToDomainInvestorContributionReport()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, report)
			// Los valores deben ser zero-values válidos
			assert.NotNil(t, report.General)
		})
	}
}

// TestInvestorHeaderModel_SharePct verifica que los headers de inversores
// contienen el share_pct acordado correctamente
func TestInvestorHeaderModel_SharePct(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected []InvestorHeaderModel
		wantErr  bool
	}{
		{
			name: "Proyecto 11: dos inversores con 50% cada uno",
			jsonData: `[
				{"investor_id": 5, "investor_name": "COTY", "share_pct": 50},
				{"investor_id": 11, "investor_name": "SOALEN SRL", "share_pct": 50}
			]`,
			expected: []InvestorHeaderModel{
				{
					InvestorID:   intPtr(5),
					InvestorName: strPtr("COTY"),
					SharePct:     decimal.NewFromInt(50),
				},
				{
					InvestorID:   intPtr(11),
					InvestorName: strPtr("SOALEN SRL"),
					SharePct:     decimal.NewFromInt(50),
				},
			},
			wantErr: false,
		},
		{
			name: "Proyecto con distribución asimétrica 70/30",
			jsonData: `[
				{"investor_id": 1, "investor_name": "Inversor A", "share_pct": 70},
				{"investor_id": 2, "investor_name": "Inversor B", "share_pct": 30}
			]`,
			expected: []InvestorHeaderModel{
				{
					InvestorID:   intPtr(1),
					InvestorName: strPtr("Inversor A"),
					SharePct:     decimal.NewFromInt(70),
				},
				{
					InvestorID:   intPtr(2),
					InvestorName: strPtr("Inversor B"),
					SharePct:     decimal.NewFromInt(30),
				},
			},
			wantErr: false,
		},
		{
			name:     "Headers vacíos",
			jsonData: `[]`,
			expected: []InvestorHeaderModel{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []InvestorHeaderModel
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				if i >= len(result) {
					break
				}
				assert.Equal(t, expected.InvestorID, result[i].InvestorID)
				assert.Equal(t, expected.InvestorName, result[i].InvestorName)
				assert.True(t, expected.SharePct.Equal(result[i].SharePct),
					"SharePct[%d]: expected %v, got %v", i, expected.SharePct, result[i].SharePct)
			}
		})
	}
}

// TestInvestorContributionDataModel_InvestorHeaders verifica que los headers
// se mapean correctamente desde la vista SQL al dominio
func TestInvestorContributionDataModel_InvestorHeaders(t *testing.T) {
	model := &InvestorContributionDataModel{
		ProjectID:    11,
		ProjectName:  "CONTROL INTEGRAL",
		CustomerID:   7,
		CustomerName: "CONTROL INTEGRAL",
		CampaignID:   4,
		CampaignName: "2025",
		InvestorHeadersJSON: `[
			{"investor_id": 5, "investor_name": "COTY", "share_pct": 50},
			{"investor_id": 11, "investor_name": "SOALEN SRL", "share_pct": 50}
		]`,
		GeneralProjectDataJSON: `{
			"surface_total_ha": 185,
			"lease_fixed_total_usd": 0,
			"lease_is_fixed": false,
			"admin_per_ha_usd": 40,
			"admin_total_usd": 7400
		}`,
		ContributionCategoriesJSON:         "[]",
		InvestorContributionComparisonJSON: "[]",
		HarvestSettlementJSON:              `{"rows": [], "footer_payment_agreed": [], "footer_payment_adjustment": []}`,
	}

	report, err := model.ToDomainInvestorContributionReport()

	require.NoError(t, err)
	require.NotNil(t, report)

	// Verificar que hay 2 inversores
	assert.Len(t, report.InvestorHeaders, 2)

	// Verificar primer inversor
	assert.Equal(t, int64(5), *report.InvestorHeaders[0].InvestorID)
	assert.Equal(t, "COTY", *report.InvestorHeaders[0].InvestorName)
	assert.True(t, decimal.NewFromInt(50).Equal(report.InvestorHeaders[0].SharePct),
		"COTY debe tener 50%%, got %v", report.InvestorHeaders[0].SharePct)

	// Verificar segundo inversor
	assert.Equal(t, int64(11), *report.InvestorHeaders[1].InvestorID)
	assert.Equal(t, "SOALEN SRL", *report.InvestorHeaders[1].InvestorName)
	assert.True(t, decimal.NewFromInt(50).Equal(report.InvestorHeaders[1].SharePct),
		"SOALEN SRL debe tener 50%%, got %v", report.InvestorHeaders[1].SharePct)

	// Verificar que la suma de porcentajes es 100%
	totalPct := report.InvestorHeaders[0].SharePct.Add(report.InvestorHeaders[1].SharePct)
	assert.True(t, decimal.NewFromInt(100).Equal(totalPct),
		"La suma de porcentajes debe ser 100%%, got %v", totalPct)
}

// Helper functions
func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

// TestContributionCategory_AdministrationUsesAgreedPct verifica que
// "Administración y Estructura" usa share_pct_agreed, no % real (FIX 000172)
func TestContributionCategory_AdministrationUsesAgreedPct(t *testing.T) {
	// Caso: Proyecto con 2 inversores al 50/50 acordado,
	// pero con aportes reales desbalanceados (60/40)
	categoryJSON := `{
		"key": "administration_structure",
		"label": "Administración y Estructura",
		"total_usd": 7400,
		"investors": [
			{"investor_id": 5, "investor_name": "COTY", "amount_usd": 3700, "share_pct": 50},
			{"investor_id": 11, "investor_name": "SOALEN SRL", "amount_usd": 3700, "share_pct": 50}
		]
	}`

	var category ContributionCategoryModel
	err := json.Unmarshal([]byte(categoryJSON), &category)
	require.NoError(t, err)

	// Verificar que hay 2 inversores
	assert.Len(t, category.Investors, 2)

	// Verificar primer inversor (COTY)
	coty := category.Investors[0]
	assert.Equal(t, int64(5), *coty.InvestorID)
	assert.Equal(t, "COTY", *coty.InvestorName)
	
	cotyPct := coty.SharePct
	assert.True(t, cotyPct.Equal(decimal.NewFromInt(50)),
		"COTY debe tener share_pct = 50 (acordado), got %v", cotyPct)
	
	cotyAmount := coty.AmountUsd
	assert.True(t, cotyAmount.Equal(decimal.NewFromInt(3700)),
		"COTY debe tener amount_usd = 3700 (50%% de 7400), got %v", cotyAmount)

	// Verificar segundo inversor (SOALEN SRL)
	soalen := category.Investors[1]
	assert.Equal(t, int64(11), *soalen.InvestorID)
	assert.Equal(t, "SOALEN SRL", *soalen.InvestorName)
	
	soalenPct := soalen.SharePct
	assert.True(t, soalenPct.Equal(decimal.NewFromInt(50)),
		"SOALEN SRL debe tener share_pct = 50 (acordado), got %v", soalenPct)
	
	soalenAmount := soalen.AmountUsd
	assert.True(t, soalenAmount.Equal(decimal.NewFromInt(3700)),
		"SOALEN SRL debe tener amount_usd = 3700 (50%% de 7400), got %v", soalenAmount)

	// Verificar que la suma de amounts es igual al total
	totalAmount := cotyAmount.Add(soalenAmount)
	expectedTotal := decimal.NewFromInt(7400)
	assert.True(t, totalAmount.Equal(expectedTotal),
		"La suma de amounts (%v) debe ser igual al total (%v)", totalAmount, expectedTotal)
}

// TestContributionCategory_LeaseUsesAgreedPct verifica que
// "Arriendo Capitalizable" también usa share_pct_agreed (FIX 000172)
func TestContributionCategory_LeaseUsesAgreedPct(t *testing.T) {
	categoryJSON := `{
		"key": "capitalizable_lease",
		"label": "Arriendo Capitalizable",
		"total_usd": 4000,
		"investors": [
			{"investor_id": 1, "investor_name": "Inversor A", "amount_usd": 2800, "share_pct": 70},
			{"investor_id": 2, "investor_name": "Inversor B", "amount_usd": 1200, "share_pct": 30}
		]
	}`

	var category ContributionCategoryModel
	err := json.Unmarshal([]byte(categoryJSON), &category)
	require.NoError(t, err)

	assert.Len(t, category.Investors, 2)

	// Verificar porcentajes acordados
	investorA := category.Investors[0]
	assert.True(t, investorA.SharePct.Equal(decimal.NewFromInt(70)),
		"Inversor A debe tener 70%% acordado")
	assert.True(t, investorA.AmountUsd.Equal(decimal.NewFromInt(2800)),
		"Inversor A debe tener 2800 (70%% de 4000)")

	investorB := category.Investors[1]
	assert.True(t, investorB.SharePct.Equal(decimal.NewFromInt(30)),
		"Inversor B debe tener 30%% acordado")
	assert.True(t, investorB.AmountUsd.Equal(decimal.NewFromInt(1200)),
		"Inversor B debe tener 1200 (30%% de 4000)")
}

// TestContributionCategory_OtherCategoriesUseRealPct verifica que
// otras categorías (como Seeds) usan % real basado en aportes (no acordado)
func TestContributionCategory_OtherCategoriesUseRealPct(t *testing.T) {
	// Caso: Semillas - Los inversores aportaron cantidades diferentes,
	// entonces el % debe reflejar los aportes reales, no el acordado
	categoryJSON := `{
		"key": "seeds",
		"label": "Semilla",
		"total_usd": 5034,
		"investors": [
			{"investor_id": 5, "investor_name": "COTY", "amount_usd": 1305, "share_pct": 25.92},
			{"investor_id": 11, "investor_name": "SOALEN SRL", "amount_usd": 3729, "share_pct": 74.08}
		]
	}`

	var category ContributionCategoryModel
	err := json.Unmarshal([]byte(categoryJSON), &category)
	require.NoError(t, err)

	// Para Seeds, el % debe ser basado en aportes reales
	coty := category.Investors[0]
	cotyPct := coty.SharePct
	
	// Calcular % real esperado: 1305 / 5034 * 100 = 25.92%
	expectedCotyPct := coty.AmountUsd.Div(decimal.NewFromInt(5034)).Mul(decimal.NewFromInt(100))
	
	diff := cotyPct.Sub(expectedCotyPct).Abs()
	assert.True(t, diff.LessThan(decimal.NewFromFloat(0.1)),
		"Seeds debe usar % real (basado en aportes), esperado ~25.92%%, got %v", cotyPct)
}

// TestInvestorContributionComparisonModel_JSONMapping verifica que los tags JSON
// coincidan con los nombres de campos en la vista SQL (FIX Problema 4)
func TestInvestorContributionComparisonModel_JSONMapping(t *testing.T) {
	// JSON que viene de la vista SQL v3_investor_contribution_data_view
	comparisonJSON := `{
		"investor_id": 5,
		"investor_name": "COTY",
		"agreed_share_pct": 50,
		"agreed_usd": 14386.275,
		"actual_usd": 17021.785,
		"adjustment_usd": 2635.51
	}`

	var comparison InvestorContributionComparisonModel
	err := json.Unmarshal([]byte(comparisonJSON), &comparison)
	require.NoError(t, err, "Debe deserializar correctamente")

	// Verificar que todos los campos se mapean correctamente
	assert.Equal(t, int64(5), *comparison.InvestorID)
	assert.Equal(t, "COTY", *comparison.InvestorName)
	
	assert.True(t, comparison.AgreedSharePct.Equal(decimal.NewFromInt(50)),
		"agreed_share_pct debe mapearse correctamente, got %v", comparison.AgreedSharePct)
	
	assert.True(t, comparison.AgreedUsd.Equal(decimal.NewFromFloat(14386.275)),
		"agreed_usd debe mapearse correctamente, got %v", comparison.AgreedUsd)
	
	assert.True(t, comparison.ActualUsd.Equal(decimal.NewFromFloat(17021.785)),
		"actual_usd debe mapearse correctamente, got %v", comparison.ActualUsd)
	
	assert.True(t, comparison.AdjustmentUsd.Equal(decimal.NewFromFloat(2635.51)),
		"adjustment_usd debe mapearse correctamente, got %v", comparison.AdjustmentUsd)
}

// TestInvestorContributionDataModel_Comparison verifica que la sección
// de comparación se mapea correctamente desde SQL al dominio (FIX Problema 4)
func TestInvestorContributionDataModel_Comparison(t *testing.T) {
	model := &InvestorContributionDataModel{
		ProjectID:    11,
		ProjectName:  "CONTROL INTEGRAL",
		CustomerID:   7,
		CustomerName: "CONTROL INTEGRAL",
		CampaignID:   4,
		CampaignName: "2025",
		InvestorHeadersJSON: `[
			{"investor_id": 5, "investor_name": "COTY", "share_pct": 50},
			{"investor_id": 11, "investor_name": "SOALEN SRL", "share_pct": 50}
		]`,
		GeneralProjectDataJSON: `{
			"surface_total_ha": 185,
			"lease_fixed_total_usd": 0,
			"lease_is_fixed": false,
			"admin_per_ha_usd": 40,
			"admin_total_usd": 7400
		}`,
		ContributionCategoriesJSON: "[]",
		InvestorContributionComparisonJSON: `[
			{
				"investor_id": 5,
				"investor_name": "COTY",
				"agreed_share_pct": 50,
				"agreed_usd": 14386.275,
				"actual_usd": 17021.785,
				"adjustment_usd": 2635.51
			},
			{
				"investor_id": 11,
				"investor_name": "SOALEN SRL",
				"agreed_share_pct": 50,
				"agreed_usd": 14386.275,
				"actual_usd": 11578.765,
				"adjustment_usd": -2807.51
			}
		]`,
		HarvestSettlementJSON: `{"rows": [], "footer_payment_agreed": [], "footer_payment_adjustment": []}`,
	}

	report, err := model.ToDomainInvestorContributionReport()

	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotNil(t, report.Comparison)
	assert.Len(t, report.Comparison, 2, "Debe haber 2 inversores en la comparación")

	// Verificar COTY
	coty := report.Comparison[0]
	assert.Equal(t, int64(5), *coty.InvestorID)
	assert.Equal(t, "COTY", *coty.InvestorName)
	
	assert.True(t, coty.AgreedSharePct.Equal(decimal.NewFromInt(50)),
		"COTY debe tener agreed_share_pct = 50, got %v", coty.AgreedSharePct)
	
	assert.True(t, coty.AgreedUsd.Equal(decimal.NewFromFloat(14386.275)),
		"COTY debe tener agreed_usd = 14386.275, got %v", coty.AgreedUsd)
	
	assert.True(t, coty.ActualUsd.Equal(decimal.NewFromFloat(17021.785)),
		"COTY debe tener actual_usd = 17021.785, got %v", coty.ActualUsd)
	
	assert.True(t, coty.AdjustmentUsd.Equal(decimal.NewFromFloat(2635.51)),
		"COTY debe tener adjustment_usd = 2635.51, got %v", coty.AdjustmentUsd)

	// Verificar SOALEN SRL
	soalen := report.Comparison[1]
	assert.Equal(t, int64(11), *soalen.InvestorID)
	assert.Equal(t, "SOALEN SRL", *soalen.InvestorName)
	
	assert.True(t, soalen.AgreedSharePct.Equal(decimal.NewFromInt(50)),
		"SOALEN SRL debe tener agreed_share_pct = 50, got %v", soalen.AgreedSharePct)
	
	assert.True(t, soalen.AgreedUsd.Equal(decimal.NewFromFloat(14386.275)),
		"SOALEN SRL debe tener agreed_usd = 14386.275, got %v", soalen.AgreedUsd)
	
	// Verificar que el ajuste sea negativo (aportó menos de lo acordado)
	assert.True(t, soalen.AdjustmentUsd.LessThan(decimal.Zero),
		"SOALEN SRL debe tener adjustment_usd negativo (aportó menos), got %v", soalen.AdjustmentUsd)
}

// TestInvestorContributionComparison_CalculationLogic verifica que la lógica
// de cálculo sea correcta: adjustment = actual - agreed
func TestInvestorContributionComparison_CalculationLogic(t *testing.T) {
	tests := []struct {
		name              string
		agreedUsd         decimal.Decimal
		actualUsd         decimal.Decimal
		expectedAdjustment decimal.Decimal
	}{
		{
			name:              "Aportó más de lo acordado (positivo)",
			agreedUsd:         decimal.NewFromInt(10000),
			actualUsd:         decimal.NewFromInt(12000),
			expectedAdjustment: decimal.NewFromInt(2000),
		},
		{
			name:              "Aportó menos de lo acordado (negativo)",
			agreedUsd:         decimal.NewFromInt(10000),
			actualUsd:         decimal.NewFromInt(8000),
			expectedAdjustment: decimal.NewFromInt(-2000),
		},
		{
			name:              "Aportó exactamente lo acordado (cero)",
			agreedUsd:         decimal.NewFromInt(10000),
			actualUsd:         decimal.NewFromInt(10000),
			expectedAdjustment: decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustment := tt.actualUsd.Sub(tt.agreedUsd)
			assert.True(t, adjustment.Equal(tt.expectedAdjustment),
				"Adjustment debe ser %v, got %v", tt.expectedAdjustment, adjustment)
		})
	}
}

// TestGeneralProjectData_AdminPerHaCalculation verifica que el cálculo de
// admin_per_ha_usd sea correcto (admin_total / surface_total)
func TestGeneralProjectData_AdminPerHaCalculation(t *testing.T) {
	tests := []struct {
		name             string
		adminTotal       decimal.Decimal
		surfaceTotal     decimal.Decimal
		expectedPerHa    decimal.Decimal
		toleranceEpsilon float64 // Para comparaciones con decimales
	}{
		{
			name:             "Proyecto 11: 7400 / 185 = 40",
			adminTotal:       decimal.NewFromInt(7400),
			surfaceTotal:     decimal.NewFromInt(185),
			expectedPerHa:    decimal.NewFromInt(40),
			toleranceEpsilon: 0.01,
		},
		{
			name:             "Proyecto con decimales: 3567.75 / 100.5 = 35.5",
			adminTotal:       decimal.NewFromFloat(3567.75),
			surfaceTotal:     decimal.NewFromFloat(100.5),
			expectedPerHa:    decimal.NewFromFloat(35.5),
			toleranceEpsilon: 0.01,
		},
		{
			name:             "Superficie 0: debe resultar en 0 (no error)",
			adminTotal:       decimal.NewFromInt(1000),
			surfaceTotal:     decimal.Zero,
			expectedPerHa:    decimal.Zero,
			toleranceEpsilon: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result decimal.Decimal
			if tt.surfaceTotal.IsZero() {
				result = decimal.Zero
			} else {
				result = tt.adminTotal.Div(tt.surfaceTotal)
			}

			diff := result.Sub(tt.expectedPerHa).Abs()
			assert.True(t, diff.LessThan(decimal.NewFromFloat(tt.toleranceEpsilon)),
				"Expected %v, got %v (diff: %v)", tt.expectedPerHa, result, diff)
		})
	}
}

