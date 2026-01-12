package dto

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestListedSupply_MarshalJSON_Rounding(t *testing.T) {
	// Crear supply con valores decimales
	supply := ListedSupply{
		ID:           1,
		Name:         "Test Supply",
		Price:        decimal.NewFromFloat(123.456),  // Debería redondearse a 123.46 (2 decimales)
		TotalUSD:     decimal.NewFromFloat(5678.901), // Debería redondearse a 5679 (entero)
		UnitID:       1,
		UnitName:     "Lts",
		CategoryName: "Test Category",
		CategoryID:   1,
		TypeName:     "Test Type",
		TypeID:       1,
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(supply)
	assert.NoError(t, err)

	// Deserializar para verificar los valores
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verificar redondeo
	assert.Equal(t, "123.46", result["price"])    // Precio U$: 2 decimales
	assert.Equal(t, "5679", result["total_usd"])  // Total U$: entero más próximo
}

func TestListSuppliesResponse_MarshalJSON_Rounding(t *testing.T) {
	// Crear respuesta con métricas
	response := ListSuppliesResponse{
		Data:        []ListedSupply{},
		TotalKg:     decimal.NewFromFloat(13500.6),  // Debería redondearse a 13501
		TotalLts:    decimal.NewFromFloat(1320.5),   // Debería redondearse a 1321
		TotalNetUSD: decimal.NewFromFloat(13454.14), // Debería redondearse a 13454
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	// Deserializar para verificar los valores
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verificar que las métricas estén redondeadas al entero más próximo
	assert.Equal(t, "13501", result["total_kg"])      // 13500.6 -> 13501
	assert.Equal(t, "1321", result["total_lts"])      // 1320.5 -> 1321
	assert.Equal(t, "13454", result["total_net_usd"]) // 13454.14 -> 13454
}

func TestListSuppliesResponse_MarshalJSON_RoundingEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		totalKg        float64
		totalLts       float64
		totalNetUSD    float64
		expectedKg     string
		expectedLts    string
		expectedNetUSD string
	}{
		{
			name:           "Redondeo hacia arriba (.5)",
			totalKg:        100.5,
			totalLts:       200.5,
			totalNetUSD:    300.5,
			expectedKg:     "101",
			expectedLts:    "201",
			expectedNetUSD: "301",
		},
		{
			name:           "Redondeo hacia abajo (.4)",
			totalKg:        100.4,
			totalLts:       200.4,
			totalNetUSD:    300.4,
			expectedKg:     "100",
			expectedLts:    "200",
			expectedNetUSD: "300",
		},
		{
			name:           "Valores enteros exactos",
			totalKg:        1000.0,
			totalLts:       2000.0,
			totalNetUSD:    3000.0,
			expectedKg:     "1000",
			expectedLts:    "2000",
			expectedNetUSD: "3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ListSuppliesResponse{
				Data:        []ListedSupply{},
				TotalKg:     decimal.NewFromFloat(tt.totalKg),
				TotalLts:    decimal.NewFromFloat(tt.totalLts),
				TotalNetUSD: decimal.NewFromFloat(tt.totalNetUSD),
			}

			jsonData, err := json.Marshal(response)
			assert.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedKg, result["total_kg"])
			assert.Equal(t, tt.expectedLts, result["total_lts"])
			assert.Equal(t, tt.expectedNetUSD, result["total_net_usd"])
		})
	}
}

func TestListedSupply_MarshalJSON_PriceRounding(t *testing.T) {
	tests := []struct {
		name          string
		price         float64
		totalUSD      float64
		expectedPrice string
		expectedTotal string
	}{
		{
			name:          "Precio con 3 decimales",
			price:         12.345,
			totalUSD:      123.456,
			expectedPrice: "12.35", // Redondea a 2 decimales
			expectedTotal: "123",   // Redondea a entero
		},
		{
			name:          "Precio exacto con 2 decimales",
			price:         10.50,
			totalUSD:      105.00,
			expectedPrice: "10.50",
			expectedTotal: "105",
		},
		{
			name:          "Valores muy pequeños",
			price:         0.006,
			totalUSD:      0.6,
			expectedPrice: "0.01", // Redondea hacia arriba
			expectedTotal: "1",    // Redondea hacia arriba
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supply := ListedSupply{
				ID:           1,
				Name:         "Test",
				Price:        decimal.NewFromFloat(tt.price),
				TotalUSD:     decimal.NewFromFloat(tt.totalUSD),
				UnitID:       1,
				UnitName:     "Lts",
				CategoryName: "Test",
				CategoryID:   1,
				TypeName:     "Test",
				TypeID:       1,
			}

			jsonData, err := json.Marshal(supply)
			assert.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedPrice, result["price"])
			assert.Equal(t, tt.expectedTotal, result["total_usd"])
		})
	}
}

