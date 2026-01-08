package dto

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

// TestBalanceCategorySerialization verifica que BalanceCategory se serialice/deserialice correctamente
func TestBalanceCategorySerialization(t *testing.T) {
	t.Run("serialize to JSON", func(t *testing.T) {
		item := BalanceItem{
			Category:    BalanceCategorySeed,
			Label:       "Semilla",
			ExecutedUSD: decimal.NewFromFloat(1000.50),
			InvestedUSD: decimal.NewFromFloat(2000.75),
			Order:       1,
		}

		jsonData, err := json.Marshal(item)
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), `"category":"SEED"`)
		assert.Contains(t, string(jsonData), `"label":"Semilla"`)
	})

	t.Run("deserialize from JSON", func(t *testing.T) {
		jsonData := `{"category":"SEED","label":"Semilla","executed_usd":"1000.50","invested_usd":"2000.75","order":1}`

		var item BalanceItem
		err := json.Unmarshal([]byte(jsonData), &item)
		assert.NoError(t, err)
		assert.Equal(t, BalanceCategorySeed, item.Category)
		assert.Equal(t, "Semilla", item.Label)

		expectedExecuted, _ := decimal.NewFromString("1000.50")
		expectedInvested, _ := decimal.NewFromString("2000.75")
		assert.Equal(t, expectedExecuted, item.ExecutedUSD)
		assert.Equal(t, expectedInvested, item.InvestedUSD)
		assert.Equal(t, 1, item.Order)
	})
}

// TestBalanceCategoryValidation verifica que la validación de categorías funcione
func TestBalanceCategoryValidation(t *testing.T) {
	t.Run("valid categories", func(t *testing.T) {
		validCategories := []BalanceCategory{
			BalanceCategoryDirectCosts,
			BalanceCategorySeed,
			BalanceCategorySupplies,
			BalanceCategoryLabors,
			BalanceCategoryLease,
			BalanceCategoryAdmin,
		}

		for _, category := range validCategories {
			assert.True(t, category.Valid(), "Category %s should be valid", category)
		}
	})

	t.Run("invalid category", func(t *testing.T) {
		invalidCategory := BalanceCategory("INVALID")
		assert.False(t, invalidCategory.Valid(), "Invalid category should not be valid")
	})
}

// TestCreateEmptyDashboardResponse verifica que use constantes del enum
func TestCreateEmptyDashboardResponse(t *testing.T) {
	response := createEmptyDashboardResponse()

	// Verificar que se usen las constantes del enum
	expectedCategories := []BalanceCategory{
		BalanceCategoryDirectCosts,
		BalanceCategorySeed,
		BalanceCategorySupplies,
		BalanceCategoryLabors,
		BalanceCategoryLease,
		BalanceCategoryAdmin,
	}

	assert.Len(t, response.ManagementBalance.Items, len(expectedCategories))

	for i, expectedCategory := range expectedCategories {
		assert.Equal(t, expectedCategory, response.ManagementBalance.Items[i].Category,
			"Item %d should have category %s", i, expectedCategory)
		assert.True(t, response.ManagementBalance.Items[i].Category.Valid(),
			"Category %s should be valid", response.ManagementBalance.Items[i].Category)
	}
}

// TestConvertManagementBalance verifica que use constantes del enum
func TestConvertManagementBalance(t *testing.T) {
	// Crear un dominio mock con datos de prueba
	domainBalance := &domain.DashboardManagementBalance{
		Summary: &domain.DashboardBalanceSummary{
			DirectCostsExecutedUSD: decimal.NewFromFloat(1000),
			DirectCostsInvestedUSD: decimal.NewFromFloat(2000),
			StockUSD:               decimal.NewFromFloat(100),
			SemillaCostUSD:         decimal.NewFromFloat(500),
			InsumosCostUSD:         decimal.NewFromFloat(300),
			LaboresCostUSD:         decimal.NewFromFloat(200),
			RentUSD:                decimal.NewFromFloat(400),
			StructureUSD:           decimal.NewFromFloat(100),
		},
	}

	// Convertir usando la función
	dtoBalance := convertManagementBalance(domainBalance)

	// Verificar que se usen las constantes del enum
	expectedCategories := []BalanceCategory{
		BalanceCategoryDirectCosts,
		BalanceCategorySeed,
		BalanceCategorySupplies,
		BalanceCategoryLabors,
		BalanceCategoryLease,
		BalanceCategoryAdmin,
	}

	assert.Len(t, dtoBalance.Items, len(expectedCategories))

	for i, expectedCategory := range expectedCategories {
		assert.Equal(t, expectedCategory, dtoBalance.Items[i].Category,
			"Item %d should have category %s", i, expectedCategory)
		assert.True(t, dtoBalance.Items[i].Category.Valid(),
			"Category %s should be valid", dtoBalance.Items[i].Category)
	}
}

// TestConvertCropIncidence verifica que use crop_id real desde la base de datos
func TestConvertCropIncidence(t *testing.T) {
	// Crear un dominio mock con datos de prueba que incluyan crop_id real
	domainIncidence := &domain.DashboardCropIncidence{
		Crops: []domain.DashboardCropBreakdown{
			{
				ID:           10, // ID real de la base de datos
				Name:         "Soja",
				Hectares:     decimal.NewFromFloat(100),
				CostUSDPerHa: decimal.NewFromFloat(500),
				IncidencePct: decimal.NewFromFloat(50),
			},
			{
				ID:           20, // ID real de la base de datos
				Name:         "Maíz",
				Hectares:     decimal.NewFromFloat(200),
				CostUSDPerHa: decimal.NewFromFloat(300),
				IncidencePct: decimal.NewFromFloat(50),
			},
		},
	}

	// Convertir usando la función
	dtoIncidence := convertCropIncidence(domainIncidence)

	// Verificar que se use el crop_id real
	assert.Len(t, dtoIncidence.Items, 2)
	assert.Equal(t, int64(10), dtoIncidence.Items[0].CropID, "Primer cultivo debe tener crop_id = 10")
	assert.Equal(t, "Soja", dtoIncidence.Items[0].Name)
	assert.Equal(t, int64(20), dtoIncidence.Items[1].CropID, "Segundo cultivo debe tener crop_id = 20")
	assert.Equal(t, "Maíz", dtoIncidence.Items[1].Name)

	// Verificar que no se use índice temporal
	for i, item := range dtoIncidence.Items {
		assert.NotEqual(t, int64(i+1), item.CropID,
			"Item %d no debe usar índice temporal %d como crop_id", i, i+1)
	}

	// Verificar el cálculo correcto del total
	// Soja: 100 ha × 500 USD/ha = 50,000 USD
	// Maíz: 200 ha × 300 USD/ha = 60,000 USD
	// Total: 110,000 USD / 300 ha = 366.67 USD/ha (aproximadamente)
	expectedTotalHectares := decimal.NewFromInt(300)                                // 100 + 200
	expectedAvgCostPerHa := decimal.NewFromInt(110000).Div(decimal.NewFromInt(300)) // 110,000 / 300

	assert.True(t, expectedTotalHectares.Equal(dtoIncidence.Total.Hectares),
		"Total hectáreas debe ser 300, obtenido: %s", dtoIncidence.Total.Hectares)
	assert.True(t, expectedAvgCostPerHa.Sub(dtoIncidence.Total.AvgCostPerHaUSD).Abs().LessThan(decimal.NewFromFloat(0.01)),
		"Costo promedio por hectárea debe ser aproximadamente 366.67 USD/ha, obtenido: %s", dtoIncidence.Total.AvgCostPerHaUSD)
}
