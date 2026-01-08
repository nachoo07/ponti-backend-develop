package dto

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetStocksResponse_MarshalJSON_Rounding(t *testing.T) {
	// Crear una respuesta con valores decimales que necesiten redondeo
	response := GetStocksResponse{
		Stocks:         []GetStockSummary{}, // Lista vacía para este test
		NetTotalUSD:    decimal.NewFromFloat(1234.56), // Debería redondearse a 1235
		TotalLiters:    decimal.NewFromFloat(567.89),  // Debería redondearse a 568
		TotalKilograms: decimal.NewFromFloat(901.23),  // Debería redondearse a 901
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	// Deserializar para verificar los valores
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verificar que los valores estén redondeados al entero más próximo
	assert.Equal(t, "1235", result["net_total_usd"])    // 1234.56 -> 1235
	assert.Equal(t, "568", result["total_liters"])      // 567.89 -> 568
	assert.Equal(t, "901", result["total_kilograms"])   // 901.23 -> 901
}

func TestGetStocksResponse_MarshalJSON_RoundingWithDecimals(t *testing.T) {
	// Probar casos específicos de redondeo
	tests := []struct {
		name           string
		netTotalUSD    float64
		totalLiters    float64
		totalKilograms float64
		expectedUSD    string
		expectedLiters string
		expectedKG     string
	}{
		{
			name:           "Redondeo hacia arriba",
			netTotalUSD:    1.5,
			totalLiters:    2.5,
			totalKilograms: 3.5,
			expectedUSD:    "2",
			expectedLiters: "3",
			expectedKG:     "4",
		},
		{
			name:           "Redondeo hacia abajo",
			netTotalUSD:    1.4,
			totalLiters:    2.4,
			totalKilograms: 3.4,
			expectedUSD:    "1",
			expectedLiters: "2",
			expectedKG:     "3",
		},
		{
			name:           "Valor entero exacto",
			netTotalUSD:    100.0,
			totalLiters:    200.0,
			totalKilograms: 300.0,
			expectedUSD:    "100",
			expectedLiters: "200",
			expectedKG:     "300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := GetStocksResponse{
				Stocks:         []GetStockSummary{},
				NetTotalUSD:    decimal.NewFromFloat(tt.netTotalUSD),
				TotalLiters:    decimal.NewFromFloat(tt.totalLiters),
				TotalKilograms: decimal.NewFromFloat(tt.totalKilograms),
			}

			jsonData, err := json.Marshal(response)
			assert.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedUSD, result["net_total_usd"])
			assert.Equal(t, tt.expectedLiters, result["total_liters"])
			assert.Equal(t, tt.expectedKG, result["total_kilograms"])
		})
	}
}
