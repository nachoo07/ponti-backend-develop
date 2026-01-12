package dto

import (
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type UpdateCloseDateRequest struct {
	CloseDate time.Time `json:"close_date"`
}

func (r *UpdateCloseDateRequest) ToDomain(updateBy *int64) *domain.Stock {
	return &domain.Stock{
		CloseDate: &r.CloseDate,
		Base: shareddomain.Base{
			UpdatedBy: updateBy,
		},
	}
}

type UpdateCloseDateResponse struct {
	Message string `json:"message"`
}

func (r *UpdateCloseDateRequest) Validate() error {
	var timeZero time.Time
	if r.CloseDate.Equal(timeZero) {
		return types.NewError(types.ErrValidation, "close_date is required", nil)
	}
	return nil
}
