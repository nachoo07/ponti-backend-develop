package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type MonthResponse struct {
	Month        string          `json:"month"`
	StartValue   decimal.Decimal `json:"start_value"`
	EndValue     decimal.Decimal `json:"end_value"`
	AverageValue decimal.Decimal `json:"average_value"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func FromDomainMonth(d *domain.DollarAverage) MonthResponse {
	return MonthResponse{
		Month:        d.Month,
		StartValue:   d.StartValue,
		EndValue:     d.EndValue,
		AverageValue: d.AvgValue,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
