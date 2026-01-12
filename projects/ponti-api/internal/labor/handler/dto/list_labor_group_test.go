package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestLaborListItem_MarshalJSON_Rounding(t *testing.T) {
	now := time.Now()
	// Crear labor con valores decimales
	labor := LaborListItem{
		WorkorderID:     1,
		WorkorderNumber: "WO-001",
		Date:            now,
		ProjectName:     "Test Project",
		FieldName:       "Test Field",
		CropName:        "Test Crop",
		LaborName:       "Test Labor",
		Contractor:      "Test Contractor",
		SurfaceHa:       decimal.NewFromFloat(10.5),
		CostHa:          decimal.NewFromFloat(100.5),
		CategoryName:    "Test Category",
		InvestorName:    "Test Investor",
		USDAvgValue:     decimal.NewFromFloat(50.5),
		NetTotal:        decimal.NewFromFloat(1234.56),  // Debería redondearse a 1235 (entero)
		TotalIVA:        decimal.NewFromFloat(259.26),   // Debería redondearse a 259 (entero)
		USDCostHa:       decimal.NewFromFloat(75.5),
		USDNetTotal:     decimal.NewFromFloat(500.5),
		InvoiceID:       1,
		InvoiceNumber:   "INV-001",
		InvoiceCompany:  "Test Company",
		InvoiceDate:     &now,
		InvoiceStatus:   "paid",
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(labor)
	assert.NoError(t, err)

	// Deserializar para verificar los valores
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verificar redondeo
	assert.Equal(t, "1235", result["net_total"]) // Total $ Neto: entero más próximo
	assert.Equal(t, "259", result["total_iva"])  // Total $ IVA: entero más próximo
}

func TestLaborListItem_MarshalJSON_RoundingEdgeCases(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		netTotal         float64
		totalIVA         float64
		expectedNetTotal string
		expectedIVA      string
	}{
		{
			name:             "Redondeo hacia arriba (.5)",
			netTotal:         100.5,
			totalIVA:         200.5,
			expectedNetTotal: "101",
			expectedIVA:      "201",
		},
		{
			name:             "Redondeo hacia abajo (.4)",
			netTotal:         100.4,
			totalIVA:         200.4,
			expectedNetTotal: "100",
			expectedIVA:      "200",
		},
		{
			name:             "Valores enteros exactos",
			netTotal:         1000.0,
			totalIVA:         2000.0,
			expectedNetTotal: "1000",
			expectedIVA:      "2000",
		},
		{
			name:             "Valores muy pequeños",
			netTotal:         0.6,
			totalIVA:         0.5,
			expectedNetTotal: "1",
			expectedIVA:      "1",
		},
		{
			name:             "Valores con múltiples decimales",
			netTotal:         1234.567,
			totalIVA:         567.891,
			expectedNetTotal: "1235",
			expectedIVA:      "568",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labor := LaborListItem{
				WorkorderID:     1,
				WorkorderNumber: "WO-001",
				Date:            now,
				ProjectName:     "Test",
				FieldName:       "Test",
				CropName:        "Test",
				LaborName:       "Test",
				Contractor:      "Test",
				SurfaceHa:       decimal.NewFromFloat(10.0),
				CostHa:          decimal.NewFromFloat(100.0),
				CategoryName:    "Test",
				InvestorName:    "Test",
				USDAvgValue:     decimal.NewFromFloat(50.0),
				NetTotal:        decimal.NewFromFloat(tt.netTotal),
				TotalIVA:        decimal.NewFromFloat(tt.totalIVA),
				USDCostHa:       decimal.NewFromFloat(75.0),
				USDNetTotal:     decimal.NewFromFloat(500.0),
				InvoiceID:       1,
				InvoiceNumber:   "INV-001",
				InvoiceCompany:  "Test",
				InvoiceDate:     &now,
				InvoiceStatus:   "paid",
			}

			jsonData, err := json.Marshal(labor)
			assert.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedNetTotal, result["net_total"])
			assert.Equal(t, tt.expectedIVA, result["total_iva"])
		})
	}
}

func TestLaborListItem_MarshalJSON_OtherFieldsPreserved(t *testing.T) {
	now := time.Now()
	labor := LaborListItem{
		WorkorderID:     1,
		WorkorderNumber: "WO-001",
		Date:            now,
		ProjectName:     "Test Project",
		FieldName:       "Test Field",
		CropName:        "Test Crop",
		LaborName:       "Test Labor",
		Contractor:      "Test Contractor",
		SurfaceHa:       decimal.NewFromFloat(10.123),
		CostHa:          decimal.NewFromFloat(100.456),
		CategoryName:    "Test Category",
		InvestorName:    "Test Investor",
		USDAvgValue:     decimal.NewFromFloat(50.789),
		NetTotal:        decimal.NewFromFloat(1234.56),
		TotalIVA:        decimal.NewFromFloat(259.26),
		USDCostHa:       decimal.NewFromFloat(75.321),
		USDNetTotal:     decimal.NewFromFloat(500.654),
		InvoiceID:       1,
		InvoiceNumber:   "INV-001",
		InvoiceCompany:  "Test Company",
		InvoiceDate:     &now,
		InvoiceStatus:   "paid",
	}

	jsonData, err := json.Marshal(labor)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verificar que otros campos mantienen 3 decimales
	assert.Equal(t, "10.123", result["surface_ha"])
	assert.Equal(t, "100.456", result["cost_ha"])
	assert.Equal(t, "50.789", result["usd_avg_value"])
	assert.Equal(t, "75.321", result["usd_cost_ha"])
	assert.Equal(t, "500.654", result["usd_net_total"])

	// Verificar que NetTotal y TotalIVA son enteros
	assert.Equal(t, "1235", result["net_total"])
	assert.Equal(t, "259", result["total_iva"])
}

