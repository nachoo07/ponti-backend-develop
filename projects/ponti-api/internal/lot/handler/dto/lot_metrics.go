package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type LotMetrics struct {
	SeededArea      decimal.Decimal `json:"seeded_area"`
	HarvestedArea   decimal.Decimal `json:"harvested_area"`
	YieldTnPerHa    decimal.Decimal `json:"yield_tn_per_ha"`
	CostPerHectare  decimal.Decimal `json:"cost_per_hectare"`
	SuperficieTotal decimal.Decimal `json:"superficie_total"`
}

// MarshalJSON aplica formato según especificación: Toneladas/ha 2 decimales, resto enteros
func (m LotMetrics) MarshalJSON() ([]byte, error) {
	aux := struct {
		SeededArea      string `json:"seeded_area"`
		HarvestedArea   string `json:"harvested_area"`
		YieldTnPerHa    string `json:"yield_tn_per_ha"`
		CostPerHectare  string `json:"cost_per_hectare"`
		SuperficieTotal string `json:"superficie_total"`
	}{
		SeededArea:      m.SeededArea.Round(0).String(),      // Sin decimales
		HarvestedArea:   m.HarvestedArea.Round(0).String(),   // Sin decimales
		YieldTnPerHa:    m.YieldTnPerHa.StringFixed(2),       // Toneladas/ha: 2 decimales
		CostPerHectare:  m.CostPerHectare.Round(0).String(),  // Sin decimales
		SuperficieTotal: m.SuperficieTotal.Round(0).String(), // Sin decimales
	}
	return json.Marshal(aux)
}

func FromDomainMetrics(d *domain.LotMetrics) LotMetrics {
	return LotMetrics{
		SeededArea:      d.SeededArea,
		HarvestedArea:   d.HarvestedArea,
		YieldTnPerHa:    d.YieldTnPerHa,
		CostPerHectare:  d.CostPerHectare,
		SuperficieTotal: d.SuperficieTotal,
	}
}
