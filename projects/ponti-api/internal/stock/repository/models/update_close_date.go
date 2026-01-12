package models

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type StockUpdate struct {
	CloseDate time.Time `gorm:"column:close_date"`
	UpdatedBy int64     `gorm:"column:updated_by"`
}

func StockUpdateCloseDateFromDomain(d *domain.Stock) *StockUpdate {
	return &StockUpdate{
		CloseDate: *d.CloseDate,
		UpdatedBy: *d.Base.UpdatedBy,
	}
}

// ToDomain convierte StockUpdateFields a dominio
func (s *StockUpdate) ToDomain() *domain.Stock {
	return &domain.Stock{
		CloseDate: &s.CloseDate,
		Base: shareddomain.Base{
			UpdatedBy: &s.UpdatedBy,
		},
	}
}
