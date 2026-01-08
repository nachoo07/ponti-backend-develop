package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type SupplyMovementExcelDTO struct {
	EntryType       string    `excel:"INGRESO"`
	ReferenceNumber string    `excel:"N° REMITO"`
	EntryDate       time.Time `excel:"FECHA"`
	InvestorName    string    `excel:"INVERSOR"`
	SupplyName      string    `excel:"INSUMO"`
	Quantity        string    `excel:"CANTIDAD"`
	SupplyUnitName  string    `excel:"UNIDAD"`
	Category        string    `excel:"RUBRO"`
	Type            string    `excel:"TIPO/CLASE"`
	ProviderName    string    `excel:"PROVEEDOR"`
	PriceUSD        float64   `excel:"PRECIO U$"`
	TotalUSD        float64   `excel:"TOTAL U$"`
}

func BuildSupplyMovementDTO(items []*domain.SupplyMovement) []SupplyMovementExcelDTO {
	out := make([]SupplyMovementExcelDTO, 0, len(items))

	for _, it := range items {

		price, _ := it.Supply.Price.Float64()
		total, _ := it.Supply.Price.Mul(it.Quantity).Float64()

		out = append(out, SupplyMovementExcelDTO{
			EntryType:       it.MovementType,
			ReferenceNumber: it.ReferenceNumber,
			EntryDate:       *it.MovementDate,
			InvestorName:    it.Investor.Name,
			SupplyName:      it.Supply.Name,
			Quantity:        decToString(it.Quantity, 2),
			SupplyUnitName:  it.Supply.UnitName,
			Category:        it.Supply.CategoryName,
			Type:            it.Supply.Type.Name,
			ProviderName:    it.Provider.Name,
			PriceUSD:        price,
			TotalUSD:        total,
		})
	}

	return out
}

func decToString(d decimal.Decimal, scale int32) string {
	if scale < 0 {
		return d.String()
	}
	return d.Round(scale).StringFixed(scale)
}
