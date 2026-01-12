package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type LaborMetrics struct {
	SurfaceHa    decimal.Decimal `json:"surface_ha"`
	NetTotalCost decimal.Decimal `json:"net_total_cost"`
	AvgCostPerHa decimal.Decimal `json:"avg_cost_per_ha"`
}

func (m LaborMetrics) MarshalJSON() ([]byte, error) {
	aux := struct {
		SurfaceHa    string `json:"surface_ha"`
		NetTotalCost string `json:"net_total_cost"`
		AvgCostPerHa string `json:"avg_cost_per_ha"`
	}{
		SurfaceHa:    m.SurfaceHa.Round(0).String(),    // Sin decimales
		NetTotalCost: m.NetTotalCost.Round(0).String(), // Sin decimales
		AvgCostPerHa: m.AvgCostPerHa.StringFixed(2),    // Costo/ha: 2 decimales
	}
	return json.Marshal(aux)
}

func FromDomainMetrics(d *domain.LaborMetrics) LaborMetrics {
	return LaborMetrics{
		SurfaceHa:    d.SurfaceHa,
		NetTotalCost: d.NetTotalCost,
		AvgCostPerHa: d.AvgCostPerHa,
	}
}
