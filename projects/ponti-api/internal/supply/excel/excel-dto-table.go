package excel

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

type SupplyTableDTO struct {
	Name         string  `excel:"NOMBRE"`
	Price        float64 `excel:"PRECIO"`
	CategoryName string  `excel:"RUBRO"`
	TypeName     string  `excel:"TIPO/CLASE"`
}

func BuildDTO(items []domain.Supply) []SupplyTableDTO {
	out := make([]SupplyTableDTO, 0, len(items))

	for _, it := range items {
		out = append(out, SupplyTableDTO{
			Name:         it.Name,
			Price:        decToFloat(it.Price, 2),
			CategoryName: it.CategoryName,
			TypeName:     it.Type.Name,
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
