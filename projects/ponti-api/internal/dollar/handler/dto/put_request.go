package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
	"github.com/shopspring/decimal"
)

type DollarAverageItem struct {
	Month        string          `json:"month" binding:"required"`
	StartValue   decimal.Decimal `json:"start_value" binding:"required"`
	EndValue     decimal.Decimal `json:"end_value" binding:"required"`
	AverageValue decimal.Decimal `json:"average_value" binding:"required"`
}

type BulkDollarAverageRequest struct {
	Year   int64               `json:"year" binding:"required"`
	Values []DollarAverageItem `json:"values" binding:"required,dive"`
}

func (b *BulkDollarAverageRequest) ToDomainSlice(projectID int64) []domain.DollarAverage {
	var out []domain.DollarAverage
	for _, item := range b.Values {
		if item.StartValue.IsZero() && item.EndValue.IsZero() {
			continue
		}
		out = append(out, domain.DollarAverage{
			ProjectID:  projectID,
			Year:       b.Year,
			Month:      item.Month,
			StartValue: item.StartValue,
			EndValue:   item.EndValue,
			AvgValue:   item.AverageValue,
		})
	}
	return out
}
