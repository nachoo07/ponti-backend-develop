// Package dto define estructuras de transporte (HTTP) para lot.
package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type LotDate struct {
	SowingDate  string `json:"sowing_date"`
	HarvestDate string `json:"harvest_date"`
	Sequence    int    `json:"sequence"`
}

type LotListElement struct {
	ID                   int64           `json:"id"`
	ProjectID            int64           `json:"project_id"`
	FieldID              int64           `json:"field_id"`
	PreviousCropID       int64           `json:"previous_crop_id"`
	CurrentCropID        int64           `json:"current_crop_id"`
	ProjectName          string          `json:"project_name"`
	FieldName            string          `json:"field_name"`
	LotName              string          `json:"lot_name"`
	PreviousCrop         string          `json:"previous_crop"`
	CurrentCrop          string          `json:"current_crop"`
	Variety              string          `json:"variety"`
	SowedArea            decimal.Decimal `json:"sowed_area"`
	Hectares             decimal.Decimal `json:"hectares"`
	Season               string          `json:"season"`
	Tons                 decimal.Decimal `json:"tons"`
	Dates                []LotDate       `json:"dates"`
	AdminCost            decimal.Decimal `json:"admin_cost"`
	UpdatedAt            *time.Time      `json:"updated_at,omitempty"`
	HarvestedArea        decimal.Decimal `json:"harvested_area"`
	HarvestDate          *time.Time      `json:"harvest_date,omitempty"`
	CostUsdPerHa         decimal.Decimal `json:"cost_usd_per_ha"`
	YieldTnPerHa         decimal.Decimal `json:"yield_tn_per_ha"`
	IncomeNetPerHa       decimal.Decimal `json:"income_net_per_ha"`
	RentPerHa            decimal.Decimal `json:"rent_per_ha"`
	ActiveTotalPerHa     decimal.Decimal `json:"active_total_per_ha"`
	OperatingResultPerHa decimal.Decimal `json:"operating_result_per_ha"`
}

// MarshalJSON implementa redondeo homogéneo para valores decimales (2) como en workorders.
func (e LotListElement) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID                   int64      `json:"id"`
		ProjectID            int64      `json:"project_id"`
		FieldID              int64      `json:"field_id"`
		PreviousCropID       int64      `json:"previous_crop_id"`
		CurrentCropID        int64      `json:"current_crop_id"`
		ProjectName          string     `json:"project_name"`
		FieldName            string     `json:"field_name"`
		LotName              string     `json:"lot_name"`
		PreviousCrop         string     `json:"previous_crop"`
		CurrentCrop          string     `json:"current_crop"`
		Variety              string     `json:"variety"`
		SowedArea            string     `json:"sowed_area"`
		Hectares             string     `json:"hectares"`
		Season               string     `json:"season"`
		Tons                 string     `json:"tons"`
		Dates                []LotDate  `json:"dates"`
		AdminCost            string     `json:"admin_cost"`
		UpdatedAt            *time.Time `json:"updated_at,omitempty"`
		HarvestedArea        string     `json:"harvested_area"`
		HarvestDate          *time.Time `json:"harvest_date,omitempty"`
		CostUsdPerHa         string     `json:"cost_usd_per_ha"`
		YieldTnPerHa         string     `json:"yield_tn_per_ha"`
		IncomeNetPerHa       string     `json:"income_net_per_ha"`
		RentPerHa            string     `json:"rent_per_ha"`
		ActiveTotalPerHa     string     `json:"active_total_per_ha"`
		OperatingResultPerHa string     `json:"operating_result_per_ha"`
	}{
		ID:                   e.ID,
		ProjectID:            e.ProjectID,
		FieldID:              e.FieldID,
		PreviousCropID:       e.PreviousCropID,
		CurrentCropID:        e.CurrentCropID,
		ProjectName:          e.ProjectName,
		FieldName:            e.FieldName,
		LotName:              e.LotName,
		PreviousCrop:         e.PreviousCrop,
		CurrentCrop:          e.CurrentCrop,
		Variety:              e.Variety,
		SowedArea:            e.SowedArea.StringFixed(2), // Superficie: 2 decimales
		Hectares:             e.Hectares.StringFixed(2),  // Superficie: 2 decimales
		Season:               e.Season,
		Tons:                 e.Tons.StringFixed(2), // Toneladas: 2 decimales
		Dates:                e.Dates,
		AdminCost:            e.AdminCost.Round(0).String(), // Sin decimales
		UpdatedAt:            e.UpdatedAt,
		HarvestedArea:        e.HarvestedArea.StringFixed(2), // Superficie: 2 decimales
		HarvestDate:          e.HarvestDate,
		CostUsdPerHa:         e.CostUsdPerHa.Round(0).String(),         // Sin decimales
		YieldTnPerHa:         e.YieldTnPerHa.StringFixed(2),            // Rendimiento: 2 decimales
		IncomeNetPerHa:       e.IncomeNetPerHa.Round(0).String(),       // Sin decimales
		RentPerHa:            e.RentPerHa.Round(0).String(),            // Sin decimales
		ActiveTotalPerHa:     e.ActiveTotalPerHa.Round(0).String(),     // Sin decimales
		OperatingResultPerHa: e.OperatingResultPerHa.Round(0).String(), // Sin decimales
	}
	return json.Marshal(aux)
}

type LotListTotals struct {
	SumSowedArea decimal.Decimal `json:"sum_sowed_area"`
	SumCost      decimal.Decimal `json:"sum_cost"`
}

func (t LotListTotals) MarshalJSON() ([]byte, error) {
	aux := struct {
		SumSowedArea string `json:"sum_sowed_area"`
		SumCost      string `json:"sum_cost"`
	}{
		SumSowedArea: t.SumSowedArea.StringFixed(2), // Superficie: 2 decimales
		SumCost:      t.SumCost.Round(0).String(),   // Sin decimales
	}
	return json.Marshal(aux)
}

type LotListResponse struct {
	PageInfo types.PageInfo   `json:"page_info"`
	Totals   LotListTotals    `json:"totals"`
	Items    []LotListElement `json:"items"`
}

// Helpers

func FromDomainListElement(d domain.LotTable) LotListElement {
	// Map fechas (máximo 3, preservando secuencia)
	dateMap := map[int]LotDate{}
	for _, dt := range d.Dates {
		sow := ""
		if dt.SowingDate != nil {
			sow = dt.SowingDate.Format("2006-01-02")
		}
		har := ""
		if dt.HarvestDate != nil {
			har = dt.HarvestDate.Format("2006-01-02")
		}
		dateMap[dt.Sequence] = LotDate{
			SowingDate:  sow,
			HarvestDate: har,
			Sequence:    dt.Sequence,
		}
	}
	// normalizamos a 3 slots (1..3)
	dates := make([]LotDate, 3)
	for seq := 1; seq <= 3; seq++ {
		if v, ok := dateMap[seq]; ok {
			dates[seq-1] = v
		} else {
			dates[seq-1] = LotDate{Sequence: seq}
		}
	}

	return LotListElement{
		ID:                   d.ID,
		ProjectID:            d.ProjectID,
		FieldID:              d.FieldID,
		PreviousCropID:       d.PreviousCropID,
		CurrentCropID:        d.CurrentCropID,
		ProjectName:          d.ProjectName,
		FieldName:            d.FieldName,
		LotName:              d.LotName,
		PreviousCrop:         d.PreviousCrop,
		CurrentCrop:          d.CurrentCrop,
		Variety:              d.Variety,
		SowedArea:            d.SowedArea,
		Hectares:             d.Hectares,
		Season:               d.Season,
		Tons:                 d.Tons,
		Dates:                dates,
		AdminCost:            d.AdminCost,
		UpdatedAt:            d.UpdatedAt,
		HarvestedArea:        d.HarvestedArea,
		HarvestDate:          d.HarvestDate,
		CostUsdPerHa:         d.CostUsdPerHa,
		YieldTnPerHa:         d.YieldTnPerHa,
		IncomeNetPerHa:       d.IncomeNetPerHa,
		RentPerHa:            d.RentPerHa,
		ActiveTotalPerHa:     d.ActiveTotalPerHa,
		OperatingResultPerHa: d.OperatingResultPerHa,
	}
}

func FromDomainList(pageInfo types.PageInfo, rows []domain.LotTable, sumSowed, sumCost decimal.Decimal) LotListResponse {
	items := make([]LotListElement, len(rows))
	for i, r := range rows {
		items[i] = FromDomainListElement(r)
	}
	return LotListResponse{
		PageInfo: pageInfo,
		Totals: LotListTotals{
			SumSowedArea: sumSowed,
			SumCost:      sumCost,
		},
		Items: items,
	}
}
