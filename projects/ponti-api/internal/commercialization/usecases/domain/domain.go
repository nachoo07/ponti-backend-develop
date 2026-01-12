package domain

import (
	"github.com/shopspring/decimal"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type CropCommercialization struct {
	ID             int64           // ID de cada registro
	ProjectID      int64           // ID del Proyecto
	CropID         int64           // ID del cultivo
	BoardPrice     decimal.Decimal // Precio en pizarra
	FreightCost    decimal.Decimal // Costo de flete
	CommercialCost decimal.Decimal // Gasto comerciales (%)
	NetPrice       decimal.Decimal // Precio neto

	shareddomain.Base
}

func (cc *CropCommercialization) CalculateNetPrice() decimal.Decimal {
	// boardPrice * commercialCost(%) / 100  = boardPricePercentage
	boardPricePercentage := cc.BoardPrice.Mul(cc.CommercialCost).Div(decimal.NewFromInt(100))

	// boardPrice - freigthCost - boardPricePercentage = NetPrice
	netPrice := cc.BoardPrice.Sub(cc.FreightCost).Sub(boardPricePercentage)

	return netPrice
}
