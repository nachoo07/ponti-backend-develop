package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type WorkorderListElement struct {
	ID                int64           `json:"id"`
	Number            string          `json:"number"`
	ProjectName       string          `json:"project_name"`
	FieldName         string          `json:"field_name"`
	LotName           string          `json:"lot_name"`
	Date              time.Time       `json:"date"`
	CropName          string          `json:"crop_name"`
	LaborName         string          `json:"labor_name"`
	LaborCategoryName string          `json:"labor_category_name"`
	TypeName          string          `json:"type_name"`
	Contractor        string          `json:"contractor"`
	SurfaceHa         decimal.Decimal `json:"surface_ha"`
	SupplyName        string          `json:"supply_name"`
	Consumption       decimal.Decimal `json:"consumption"`
	CategoryName      string          `json:"category_name"`
	Dose              decimal.Decimal `json:"dose"`
	CostPerHa         decimal.Decimal `json:"cost_per_ha"`
	UnitPrice         decimal.Decimal `json:"unit_price"`
	TotalCost         decimal.Decimal `json:"total_cost"`
}

// MarshalJSON asegura 2 decimales en todos los campos decimal de salida
func (w WorkorderListElement) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID                int64     `json:"id"`
		Number            string    `json:"number"`
		ProjectName       string    `json:"project_name"`
		FieldName         string    `json:"field_name"`
		LotName           string    `json:"lot_name"`
		Date              time.Time `json:"date"`
		CropName          string    `json:"crop_name"`
		LaborName         string    `json:"labor_name"`
		LaborCategoryName string    `json:"labor_category_name"`
		TypeName          string    `json:"type_name"`
		Contractor        string    `json:"contractor"`
		SurfaceHa         string    `json:"surface_ha"`
		SupplyName        string    `json:"supply_name"`
		Consumption       string    `json:"consumption"`
		CategoryName      string    `json:"category_name"`
		Dose              string    `json:"dose"`
		CostPerHa         string    `json:"cost_per_ha"`
		UnitPrice         string    `json:"unit_price"`
		TotalCost         string    `json:"total_cost"`
	}{
		ID:                w.ID,
		Number:            w.Number,
		ProjectName:       w.ProjectName,
		FieldName:         w.FieldName,
		LotName:           w.LotName,
		Date:              w.Date,
		CropName:          w.CropName,
		LaborName:         w.LaborName,
		LaborCategoryName: w.LaborCategoryName,
		TypeName:          w.TypeName,
		Contractor:        w.Contractor,
		SurfaceHa:         w.SurfaceHa.StringFixed(2), // Superficie: 2 decimales
		SupplyName:        w.SupplyName,
		Consumption:       w.Consumption.StringFixed(2), // 2 decimales
		CategoryName:      w.CategoryName,
		Dose:              w.Dose.StringFixed(2),         // Dosis: 2 decimales
		CostPerHa:         w.CostPerHa.StringFixed(2),    // Costo/ha: 2 decimales
		UnitPrice:         w.UnitPrice.StringFixed(2),    // Precio/ha: 2 decimales
		TotalCost:         w.TotalCost.Round(0).String(), // Total costo: sin decimales
	}
	return json.Marshal(aux)
}

type WorkorderListResponse struct {
	PageInfo types.PageInfo         `json:"page_info"`
	Items    []WorkorderListElement `json:"items"`
}

func FromDomainListElement(d *domain.WorkorderListElement) *WorkorderListElement {
	return &WorkorderListElement{
		ID:                d.ID,
		Number:            d.Number,
		ProjectName:       d.ProjectName,
		FieldName:         d.FieldName,
		LotName:           d.LotName,
		Date:              d.Date,
		CropName:          d.CropName,
		LaborName:         d.LaborName,
		LaborCategoryName: d.LaborCategoryName,
		TypeName:          d.TypeName,
		Contractor:        d.Contractor,
		SurfaceHa:         d.SurfaceHa,
		SupplyName:        d.SupplyName,
		Consumption:       d.Consumption,
		CategoryName:      d.CategoryName,
		Dose:              d.Dose,
		CostPerHa:         d.CostPerHa,
		UnitPrice:         d.UnitPrice,
		TotalCost:         d.TotalCost,
	}
}

func FromDomainList(pageInfo types.PageInfo, list []domain.WorkorderListElement) WorkorderListResponse {
	items := make([]WorkorderListElement, len(list))
	for i, d := range list {
		items[i] = *FromDomainListElement(&d)
	}
	return WorkorderListResponse{PageInfo: pageInfo, Items: items}
}
