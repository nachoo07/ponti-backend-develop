package dto

import (
	"github.com/shopspring/decimal"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type UpdateRealStockRequest struct {
	RealStockUnits decimal.Decimal `json:"real_stock_units"`
}

func (r *UpdateRealStockRequest) ToDomain(updatedBy *int64) *domain.Stock {
	return &domain.Stock{
		RealStockUnits: r.RealStockUnits,
		Base: shareddomain.Base{
			UpdatedBy: updatedBy,
		},
	}
}

type UpdateRealStockResponse struct {
	Message string `json:"message"`
}

func NewUpdateRealStockResponse(message string) *UpdateRealStockResponse {
	return &UpdateRealStockResponse{
		Message: message,
	}
}
