package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	"github.com/shopspring/decimal"
)

type LotTableExcelDTO struct {
	ProjectName          string     `excel:"PROYECTO"`
	FieldName            string     `excel:"CAMPO"`
	LotName              string     `excel:"LOTES"`
	PreviousCrop         string     `excel:"CULTIVO ANTERIOR"`
	CurrentCrop          string     `excel:"CULTIVO ACTUAL"`
	Variety              string     `excel:"VARIEDAD"`
	SowedArea            float64    `excel:"SUP. SIEMBRA"`
	SowingDate           *time.Time `excel:"FECHA SIEMBRA"`
	CostUsdPerHa         float64    `excel:"COSTO U$/HA"`
	HarvestedArea        float64    `excel:"SUP. COSECHA"`
	HarvestedDate        *time.Time `excel:"FECHA COSECHA"`
	Tons                 float64    `excel:"TONELADAS"`
	YieldTnPerHa         float64    `excel:"RENDIMIENTO"`
	IncomeNetPerHa       float64    `excel:"INGRESO NETO/HA"`
	RentPerHa            float64    `excel:"ARRIENDO/HA"`
	AdminCost            float64    `excel:"ADMIN. PROYECTO/HA"`
	ActiveTotalPerHa     float64    `excel:"ACTIVO TOTAL/HA"`
	OperatingResultPerHa float64    `excel:"RESULTADO OPERATIVO"`
}

func BuildLotExcelDTO(items []domain.LotTable) []LotTableExcelDTO {
	out := make([]LotTableExcelDTO, 0, len(items))

	for _, it := range items {
		var sowingDate *time.Time
		var harvestedDate *time.Time

		for _, dt := range it.Dates {
			if sowingDate == nil && dt.SowingDate != nil {
				sowingDate = dt.SowingDate
			}
			if harvestedDate == nil && dt.HarvestDate != nil {
				harvestedDate = dt.HarvestDate
			}
			if sowingDate != nil && harvestedDate != nil {
				break
			}
		}

		out = append(out, LotTableExcelDTO{
			ProjectName:          it.ProjectName,
			FieldName:            it.FieldName,
			LotName:              it.LotName,
			PreviousCrop:         it.PreviousCrop,
			CurrentCrop:          it.CurrentCrop,
			Variety:              it.Variety,
			SowedArea:            decToFloat(it.SowedArea, 2),
			SowingDate:           sowingDate,
			CostUsdPerHa:         decToFloat(it.CostUsdPerHa, 2),
			HarvestedArea:        decToFloat(it.HarvestedArea, 2),
			HarvestedDate:        harvestedDate,
			Tons:                 decToFloat(it.Tons, 2),
			YieldTnPerHa:         decToFloat(it.YieldTnPerHa, 2),
			IncomeNetPerHa:       decToFloat(it.IncomeNetPerHa, 2),
			RentPerHa:            decToFloat(it.RentPerHa, 2),
			AdminCost:            decToFloat(it.AdminCost, 2),
			ActiveTotalPerHa:     decToFloat(it.ActiveTotalPerHa, 2),
			OperatingResultPerHa: decToFloat(it.OperatingResultPerHa, 2),
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
