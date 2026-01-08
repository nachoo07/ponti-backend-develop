package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	decimal "github.com/shopspring/decimal"
)

type CropCommercialization struct {
	ID             int64           `json:"id"`
	CropID         int64           `json:"crop_id" binding:"required,gt=0"`
	BoardPrice     decimal.Decimal `json:"board_price" binding:"required"`
	FreightCost    decimal.Decimal `json:"freight_cost" binding:"required"`
	CommercialCost decimal.Decimal `json:"commercial_cost" binding:"required"`
}

type BulkCommercializationRequest struct {
	Values []CropCommercialization `json:"values" binding:"required,dive"`
}

func (b *BulkCommercializationRequest) ToDomainSlice(projecID int64, userID int64) []domain.CropCommercialization {
	out := make([]domain.CropCommercialization, len(b.Values))
	for i, item := range b.Values {
		out[i] = domain.CropCommercialization{
			ID:             item.ID,
			ProjectID:      projecID,
			CropID:         item.CropID,
			BoardPrice:     item.BoardPrice,
			FreightCost:    item.FreightCost,
			CommercialCost: item.CommercialCost,
			Base: shareddomain.Base{
				CreatedBy: &userID,
				UpdatedBy: &userID,
			},
		}
	}
	return out
}
