package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
	"github.com/shopspring/decimal"
)

type WorkorderExcelDto struct {
	Number       string    `excel:"NUMERO DE ORDEN"`
	ProjectName  string    `excel:"PROYECTO"`
	FieldName    string    `excel:"CAMPO"`
	LotName      string    `excel:"LOTE"`
	Date         time.Time `excel:"FECHA"`
	CropName     string    `excel:"CULTIVO"`
	LaborName    string    `excel:"LABOR"`
	TypeName     string    `excel:"TIPO/CLASE"`
	Contractor   string    `excel:"CONTRATISTA"`
	SurfaceHa    float64   `excel:"SUPERFICIE"`
	SupplyName   string    `excel:"INSUMO"`
	Consumption  float64   `excel:"CONSUMO"`
	CategoryName string    `excel:"RUBRO"`
	Dose         float64   `excel:"DOSIS"`
	CostPerHa    float64   `excel:"COST U$/HA"`
	UnitPrice    float64   `excel:"PRECIO UNIDAD"`
	TotalCost    float64   `excel:"TOTAL COSTO"`
}

func BuildWorkorderExcelDTO(items []domain.WorkorderListElement) []WorkorderExcelDto {
	out := make([]WorkorderExcelDto, 0, len(items)+1) // +1 para la fila de totales

	// Variables para acumular totales
	var totalSurfaceHa decimal.Decimal
	var totalConsumption decimal.Decimal
	var totalCost decimal.Decimal

	for _, it := range items {
		// Acumular totales
		totalSurfaceHa = totalSurfaceHa.Add(it.SurfaceHa)
		totalConsumption = totalConsumption.Add(it.Consumption)
		totalCost = totalCost.Add(it.TotalCost)

		out = append(out, WorkorderExcelDto{
			Number:       it.Number,
			ProjectName:  it.ProjectName,
			FieldName:    it.FieldName,
			LotName:      it.LotName,
			Date:         it.Date,
			CropName:     it.CropName,
			LaborName:    it.LaborName,
			TypeName:     it.TypeName,
			Contractor:   it.Contractor,
			SurfaceHa:    decToFloat(it.SurfaceHa, 2),
			SupplyName:   it.SupplyName,
			Consumption:  decToFloat(it.Consumption, 0),
			CategoryName: it.CategoryName,
			Dose:         decToFloat(it.Dose, 2),
			CostPerHa:    decToFloat(it.CostPerHa, 2),
			UnitPrice:    decToFloat(it.UnitPrice, 2),
			TotalCost:    decToFloat(it.TotalCost, 2),
		})
	}

	// Agregar fila de totales al final
	if len(items) > 0 {
		out = append(out, WorkorderExcelDto{
			Number:       "TOTAL",
			ProjectName:  "",
			FieldName:    "",
			LotName:      "",
			Date:         time.Time{},
			CropName:     "",
			LaborName:    "",
			TypeName:     "",
			Contractor:   "",
			SurfaceHa:    decToFloat(totalSurfaceHa, 2),
			SupplyName:   "",
			Consumption:  decToFloat(totalConsumption, 0),
			CategoryName: "",
			Dose:         0,
			CostPerHa:    0,
			UnitPrice:    0,
			TotalCost:    decToFloat(totalCost, 2),
		})
	}

	return out
}

// helper para convertir decimal en float64 con redondeo opcional
func decToFloat(d decimal.Decimal, scale int32) float64 {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return f
}

// helper para obtener string con decimales fijos
func decToString(d decimal.Decimal, scale int32) string {
	if scale < 0 {
		return d.String()
	}
	return d.Round(scale).StringFixed(scale)
}
