package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	investorAPIKey      = "abc123secreta"
	investorUserID      = "123"
	investorProjectID   = 12
	investorProjectName = "JUJUY PRUEBA"
)

// GeneralDataResponse estructura para los datos generales del proyecto (Card Superficie)
type GeneralDataResponse struct {
	SurfaceTotalHa string `json:"surface_total_ha"`
	LeaseFixedUsd  string `json:"lease_fixed_usd"`
	LeaseIsFixed   bool   `json:"lease_is_fixed"`
	AdminPerHaUsd  string `json:"admin_per_ha_usd"`
	AdminTotalUsd  string `json:"admin_total_usd"`
}

// InvestorHeaderResponse estructura para los headers de inversores
type InvestorHeaderResponse struct {
	InvestorID   int64  `json:"investor_id"`
	InvestorName string `json:"investor_name"`
	SharePct     string `json:"share_pct"`
}

// ContributionInvestorResponse estructura para inversores dentro de una categoría
type ContributionInvestorResponse struct {
	InvestorID   int64  `json:"investor_id"`
	InvestorName string `json:"investor_name"`
	AmountUsd    string `json:"amount_usd"`
	SharePct     string `json:"share_pct"`
}

// ContributionCategoryResponse estructura para una categoría de contribución
type ContributionCategoryResponse struct {
	Key       string                         `json:"key"`
	Label     string                         `json:"label"`
	TotalUsd  string                         `json:"total_usd"`
	Investors []ContributionInvestorResponse `json:"investors"`
}

// ComparisonResponse estructura para la sección de comparación (Acordado vs Real)
type ComparisonResponse struct {
	InvestorID     int64  `json:"investor_id"`
	InvestorName   string `json:"investor_name"`
	AgreedSharePct string `json:"agreed_share_pct"`
	AgreedUsd      string `json:"agreed_usd"`
	ActualUsd      string `json:"actual_usd"`
	AdjustmentUsd  string `json:"adjustment_usd"`
}

// InvestorContributionResponse estructura simplificada del reporte
type InvestorContributionResponse struct {
	ProjectID       int64                          `json:"project_id"`
	ProjectName     string                         `json:"project_name"`
	InvestorHeaders []InvestorHeaderResponse       `json:"investor_headers"`
	General         GeneralDataResponse            `json:"general"`
	Contributions   []ContributionCategoryResponse `json:"contributions"`
	Comparison      []ComparisonResponse           `json:"comparison"`
}

// TestInvestorContribution_GeneralData_E2E verifica que los datos generales
// del proyecto (Card Superficie) se devuelven correctamente desde el endpoint
func TestInvestorContribution_GeneralData_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	// Crear request
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	// Ejecutar request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verificar status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "El endpoint debe devolver 200 OK")

	// Parsear response
	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err, "La respuesta debe ser JSON válido")

	// Verificar datos básicos
	assert.Equal(t, int64(investorProjectID), report.ProjectID, "Project ID debe ser 12")
	assert.Equal(t, investorProjectName, report.ProjectName, "Project name debe ser JUJUY PRUEBA")

	// Verificar datos generales (Card Superficie) - FIX de migración 000171
	t.Run("Card Superficie tiene valores correctos", func(t *testing.T) {
		general := report.General

		// Surface total
		surfaceTotal, err := decimal.NewFromString(general.SurfaceTotalHa)
		require.NoError(t, err, "surface_total_ha debe ser un número válido")
		assert.True(t, surfaceTotal.Equal(decimal.NewFromInt(1189)),
			"surface_total_ha debe ser 1189, got %s", general.SurfaceTotalHa)

		// Admin total
		adminTotal, err := decimal.NewFromString(general.AdminTotalUsd)
		require.NoError(t, err, "admin_total_usd debe ser un número válido")
		assert.True(t, adminTotal.Equal(decimal.NewFromInt(49938)),
			"admin_total_usd debe ser 49938, got %s", general.AdminTotalUsd)

		// Admin per ha
		adminPerHa, err := decimal.NewFromString(general.AdminPerHaUsd)
		require.NoError(t, err, "admin_per_ha_usd debe ser un número válido")
		assert.True(t, adminPerHa.Equal(decimal.NewFromInt(42)),
			"admin_per_ha_usd debe ser 42, got %s", general.AdminPerHaUsd)

		// Lease fixed
		leaseFixed, err := decimal.NewFromString(general.LeaseFixedUsd)
		require.NoError(t, err, "lease_fixed_usd debe ser un número válido")
		assert.True(t, leaseFixed.Equal(decimal.NewFromInt(119770)),
			"lease_fixed_usd debe ser 119770, got %s", general.LeaseFixedUsd)

		// Lease is fixed
		assert.True(t, general.LeaseIsFixed,
			"lease_is_fixed debe ser true para proyecto 12")
	})
}

// TestInvestorContribution_GeneralData_NotZero_E2E verifica que los valores
// NO son cero cuando deberían tener datos (regresión de bug previo)
func TestInvestorContribution_GeneralData_NotZero_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// REGRESIÓN: Antes de la migración 000171, todos estos valores eran "0"
	t.Run("Verificar que NO son cero (regresión)", func(t *testing.T) {
		assert.NotEqual(t, "0", report.General.SurfaceTotalHa,
			"surface_total_ha NO debe ser 0")
		assert.NotEqual(t, "0", report.General.AdminTotalUsd,
			"admin_total_usd NO debe ser 0")
		assert.NotEqual(t, "0", report.General.AdminPerHaUsd,
			"admin_per_ha_usd NO debe ser 0")
	})
}

// TestInvestorContribution_GeneralData_CalculationConsistency_E2E verifica
// que los cálculos sean consistentes (admin_per_ha = admin_total / surface)
func TestInvestorContribution_GeneralData_CalculationConsistency_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Consistencia de cálculos: admin_per_ha = admin_total / surface", func(t *testing.T) {
		surfaceTotal, err := decimal.NewFromString(report.General.SurfaceTotalHa)
		require.NoError(t, err)

		adminTotal, err := decimal.NewFromString(report.General.AdminTotalUsd)
		require.NoError(t, err)

		adminPerHa, err := decimal.NewFromString(report.General.AdminPerHaUsd)
		require.NoError(t, err)

		// Calcular admin_per_ha esperado
		var expectedAdminPerHa decimal.Decimal
		if surfaceTotal.IsZero() {
			expectedAdminPerHa = decimal.Zero
		} else {
			expectedAdminPerHa = adminTotal.Div(surfaceTotal)
		}

		// Permitir pequeña diferencia por redondeo (< 0.01)
		diff := adminPerHa.Sub(expectedAdminPerHa).Abs()
		assert.True(t, diff.LessThan(decimal.NewFromFloat(0.01)),
			"admin_per_ha (%s) debe ser consistente con admin_total / surface (%s), diff: %s",
			adminPerHa, expectedAdminPerHa, diff)
	})
}

// TestInvestorContribution_InvestorHeaders_E2E verifica que los títulos de inversores
// muestran el porcentaje acordado correctamente (FIX de migración 000171)
func TestInvestorContribution_InvestorHeaders_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Investor headers tienen porcentajes correctos", func(t *testing.T) {
		expectedHeaders := []struct {
			id   int64
			name string
			pct  int64
		}{
			{7, "agro lajitas", 47},
			{13, "olega", 47},
			{14, "vedoya", 6},
		}

		assert.Len(t, report.InvestorHeaders, len(expectedHeaders), "Debe haber 3 inversores")

		totalPct := decimal.Zero
		for i, expected := range expectedHeaders {
			header := report.InvestorHeaders[i]
			assert.Equal(t, expected.id, header.InvestorID)
			assert.Equal(t, expected.name, header.InvestorName)

			sharePct, err := decimal.NewFromString(header.SharePct)
			require.NoError(t, err, "share_pct debe ser un número válido")
			assert.True(t, sharePct.Equal(decimal.NewFromInt(expected.pct)),
				"El inversor %s debe tener %d%%, got %s", header.InvestorName, expected.pct, header.SharePct)

			totalPct = totalPct.Add(sharePct)
		}

		assert.True(t, totalPct.Equal(decimal.NewFromInt(100)),
			"La suma de porcentajes debe ser 100%%, got %s", totalPct)
	})
}

// TestInvestorContribution_InvestorHeaders_NotZero_E2E verifica que los porcentajes
// NO son cero (regresión de bug previo - Problema 1)
func TestInvestorContribution_InvestorHeaders_NotZero_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// REGRESIÓN: Antes de la migración 000171, todos los share_pct eran "0"
	t.Run("Verificar que share_pct NO son cero (regresión Problema 1)", func(t *testing.T) {
		for i, header := range report.InvestorHeaders {
			assert.NotEqual(t, "0", header.SharePct,
				"Inversor %d (%s) NO debe tener share_pct = 0", i, header.InvestorName)

			pct, err := decimal.NewFromString(header.SharePct)
			require.NoError(t, err)
			assert.True(t, pct.GreaterThan(decimal.Zero),
				"Inversor %s debe tener porcentaje > 0", header.InvestorName)
		}
	})
}

// TestInvestorContribution_AdministrationUsesAgreedPct_E2E verifica que
// "Administración y Estructura" usa el % acordado, no el % real (FIX 000172)
func TestInvestorContribution_AdministrationUsesAgreedPct_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// Buscar la categoría "Administración y Estructura"
	var adminCategory *ContributionCategoryResponse
	for _, cat := range report.Contributions {
		if cat.Key == "administration" {
			adminCategory = &cat
			break
		}
	}

	require.NotNil(t, adminCategory, "Debe existir la categoría 'administration'")

	t.Run("Administración y Estructura usa % acordado (50/50)", func(t *testing.T) {
		assert.Equal(t, "Administración y Estructura", adminCategory.Label)

		totalUsd, err := decimal.NewFromString(adminCategory.TotalUsd)
		require.NoError(t, err)
		assert.True(t, totalUsd.Equal(decimal.NewFromInt(49938)),
			"Total debe ser 49938, got %s", adminCategory.TotalUsd)

		// Deben haber 3 inversores
		assert.Len(t, adminCategory.Investors, 3, "Debe haber 3 inversores")

		expectedInvestors := []struct {
			id    int64
			name  string
			share int64
			usd   int64
		}{
			{7, "agro lajitas", 47, 23471},
			{13, "olega", 47, 23471},
			{14, "vedoya", 6, 2996},
		}

		sumAmounts := decimal.Zero
		for i, expected := range expectedInvestors {
			investor := adminCategory.Investors[i]
			assert.Equal(t, expected.id, investor.InvestorID)
			assert.Equal(t, expected.name, investor.InvestorName)

			sharePct, err := decimal.NewFromString(investor.SharePct)
			require.NoError(t, err)
			assert.True(t, sharePct.Equal(decimal.NewFromInt(expected.share)),
				"Inversor %s debe tener %d%% (acordado), got %s", investor.InvestorName, expected.share, investor.SharePct)

			amount, err := decimal.NewFromString(investor.AmountUsd)
			require.NoError(t, err)
			assert.True(t, amount.Equal(decimal.NewFromInt(expected.usd)),
				"Inversor %s debe tener %d, got %s", investor.InvestorName, expected.usd, investor.AmountUsd)

			sumAmounts = sumAmounts.Add(amount)
		}

		assert.True(t, sumAmounts.Equal(totalUsd),
			"La suma de amounts (%s) debe ser igual al total (%s)", sumAmounts, totalUsd)
	})
}

// TestInvestorContribution_AdministrationNotRealPct_E2E verifica que
// Administración NO usa el % real basado en aportes (regresión de Problema 3)
func TestInvestorContribution_AdministrationNotRealPct_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// Buscar la categoría "Administración y Estructura"
	var adminCategory *ContributionCategoryResponse
	for _, cat := range report.Contributions {
		if cat.Key == "administration" {
			adminCategory = &cat
			break
		}
	}

	require.NotNil(t, adminCategory)

	// REGRESIÓN: Antes de la migración 000172, los porcentajes se calculaban con aportes reales.
	// Ahora deben respetar los porcentajes acordados para Agro Lajitas / Olega / Vedoya.

	t.Run("Verificar que NO usa % real (regresión Problema 3)", func(t *testing.T) {
		expectedShares := map[string]int64{
			"agro lajitas": 47,
			"olega":        47,
			"vedoya":       6,
		}

		for _, investor := range adminCategory.Investors {
			pct, err := decimal.NewFromString(investor.SharePct)
			require.NoError(t, err)

			expectedPct, ok := expectedShares[investor.InvestorName]
			require.True(t, ok, "No se esperaba el inversor %s en administración", investor.InvestorName)

			assert.True(t, pct.Equal(decimal.NewFromInt(expectedPct)),
				"Inversor %s debe tener %d%% (%s), got %s",
				investor.InvestorName, expectedPct, "acordado", investor.SharePct)
		}
	})
}

// TestInvestorContribution_LeaseUsesAgreedPct_E2E verifica que
// "Arriendo Capitalizable" también usa % acordado (FIX 000172)
func TestInvestorContribution_LeaseUsesAgreedPct_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// Buscar la categoría "Arriendo Capitalizable"
	var leaseCategory *ContributionCategoryResponse
	for _, cat := range report.Contributions {
		if cat.Key == "rent_capitalizable" {
			leaseCategory = &cat
			break
		}
	}

	require.NotNil(t, leaseCategory, "Debe existir la categoría 'rent_capitalizable'")

	t.Run("Arriendo Capitalizable usa % acordado", func(t *testing.T) {
		assert.Equal(t, "Arriendo Capitalizable", leaseCategory.Label)

		expectedShares := map[string]int64{
			"agro lajitas": 50,
			"olega":        50,
			"vedoya":       0,
		}

		for _, investor := range leaseCategory.Investors {
			pct, err := decimal.NewFromString(investor.SharePct)
			require.NoError(t, err)

			expectedPct, ok := expectedShares[investor.InvestorName]
			require.True(t, ok, "No se esperaba el inversor %s en arriendo", investor.InvestorName)

			assert.True(t, pct.Equal(decimal.NewFromInt(expectedPct)),
				"Inversor %s debe tener %d%% (acordado), got %s", investor.InvestorName, expectedPct, investor.SharePct)
		}
	})
}

// TestInvestorContribution_ComparisonHasData_E2E verifica que la sección
// de comparación (Acordado vs Real) devuelve datos correctos (FIX Problema 4)
func TestInvestorContribution_ComparisonHasData_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Comparación tiene datos correctos", func(t *testing.T) {
		require.NotNil(t, report.Comparison, "Debe existir la sección 'comparison'")
		assert.Len(t, report.Comparison, 3, "Debe haber 3 inversores en la comparación")

		expectedComparison := []struct {
			id        int64
			name      string
			sharePct  int64
			minAgreed int64
			minActual int64
		}{
			{7, "agro lajitas", 47, 228000, 231000},
			{13, "olega", 47, 228000, 231000},
			{14, "vedoya", 6, 20000, 20000},
		}

		for i, expected := range expectedComparison {
			row := report.Comparison[i]
			assert.Equal(t, expected.id, row.InvestorID)
			assert.Equal(t, expected.name, row.InvestorName)

			sharePct, err := decimal.NewFromString(row.AgreedSharePct)
			require.NoError(t, err)
			assert.True(t, sharePct.Equal(decimal.NewFromInt(expected.sharePct)),
				"%s debe tener agreed_share_pct = %d, got %s", row.InvestorName, expected.sharePct, row.AgreedSharePct)

			agreed, err := decimal.NewFromString(row.AgreedUsd)
			require.NoError(t, err)
			assert.True(t, agreed.GreaterThan(decimal.Zero),
				"%s debe tener agreed_usd > 0, got %s", row.InvestorName, row.AgreedUsd)

			actual, err := decimal.NewFromString(row.ActualUsd)
			require.NoError(t, err)
			assert.True(t, actual.GreaterThan(decimal.Zero),
				"%s debe tener actual_usd > 0, got %s", row.InvestorName, row.ActualUsd)
		}
	})
}

// TestInvestorContribution_ComparisonNotZero_E2E verifica que los valores
// NO son cero (regresión de Problema 4: antes del fix todos eran "0")
func TestInvestorContribution_ComparisonNotZero_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// REGRESIÓN: Antes del fix, todos los valores eran "0"
	t.Run("Verificar que NO son cero (regresión Problema 4)", func(t *testing.T) {
		for _, comparison := range report.Comparison {
			assert.NotEqual(t, "0", comparison.AgreedSharePct,
				"Inversor %s NO debe tener agreed_share_pct = 0", comparison.InvestorName)
			assert.NotEqual(t, "0", comparison.AgreedUsd,
				"Inversor %s NO debe tener agreed_usd = 0", comparison.InvestorName)
			assert.NotEqual(t, "0", comparison.ActualUsd,
				"Inversor %s NO debe tener actual_usd = 0", comparison.InvestorName)
		}
	})
}

// TestInvestorContribution_ComparisonCalculation_E2E verifica que el cálculo
// de ajuste sea correcto: adjustment = actual - agreed
func TestInvestorContribution_ComparisonCalculation_E2E(t *testing.T) {
	projectID := investorProjectID

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Verificar cálculo de ajuste: adjustment = actual - agreed", func(t *testing.T) {
		for _, comparison := range report.Comparison {
			agreed, err := decimal.NewFromString(comparison.AgreedUsd)
			require.NoError(t, err)

			actual, err := decimal.NewFromString(comparison.ActualUsd)
			require.NoError(t, err)

			adjustment, err := decimal.NewFromString(comparison.AdjustmentUsd)
			require.NoError(t, err)

			expectedAdjustment := actual.Sub(agreed)

			// Permitir pequeña diferencia por redondeo (hasta 1 USD)
			diff := adjustment.Sub(expectedAdjustment).Abs()
			assert.True(t, diff.LessThanOrEqual(decimal.NewFromInt(1)),
				"Inversor %s: adjustment (%s) debe ser actual - agreed (%s), diff: %s",
				comparison.InvestorName, adjustment, expectedAdjustment, diff)
		}
	})
}
