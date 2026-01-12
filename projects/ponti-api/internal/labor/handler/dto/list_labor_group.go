package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type LaborListItem struct {
	WorkorderID     int64           `json:"workorder_id"`
	WorkorderNumber string          `json:"workorder_number"`
	Date            time.Time       `json:"date"`
	ProjectName     string          `json:"project_name"`
	FieldName       string          `json:"field_name"`
	CropName        string          `json:"crop_name"`
	LaborName       string          `json:"labor_name"`
	Contractor      string          `json:"contractor"`
	SurfaceHa       decimal.Decimal `json:"surface_ha"`
	CostHa          decimal.Decimal `json:"cost_ha"`
	CategoryName    string          `json:"category_name"`
	InvestorName    string          `json:"investor_name"`
	USDAvgValue     decimal.Decimal `json:"usd_avg_value"`
	NetTotal        decimal.Decimal `json:"net_total"`        // Total $ Neto
	TotalIVA        decimal.Decimal `json:"total_iva"`        // Total $ IVA
	USDCostHa       decimal.Decimal `json:"usd_cost_ha"`
	USDNetTotal     decimal.Decimal `json:"usd_net_total"`
	InvoiceID       int64           `json:"invoice_id"`
	InvoiceNumber   string          `json:"invoice_number"`
	InvoiceCompany  string          `json:"invoice_company"`
	InvoiceDate     *time.Time      `json:"invoice_date"`
	InvoiceStatus   string          `json:"invoice_status"`
}

// MarshalJSON aplica redondeo: Total $ Neto y Total $ IVA al entero más próximo
func (l LaborListItem) MarshalJSON() ([]byte, error) {
	aux := struct {
		WorkorderID     int64      `json:"workorder_id"`
		WorkorderNumber string     `json:"workorder_number"`
		Date            time.Time  `json:"date"`
		ProjectName     string     `json:"project_name"`
		FieldName       string     `json:"field_name"`
		CropName        string     `json:"crop_name"`
		LaborName       string     `json:"labor_name"`
		Contractor      string     `json:"contractor"`
		SurfaceHa       string     `json:"surface_ha"`
		CostHa          string     `json:"cost_ha"`
		CategoryName    string     `json:"category_name"`
		InvestorName    string     `json:"investor_name"`
		USDAvgValue     string     `json:"usd_avg_value"`
		NetTotal        string     `json:"net_total"`
		TotalIVA        string     `json:"total_iva"`
		USDCostHa       string     `json:"usd_cost_ha"`
		USDNetTotal     string     `json:"usd_net_total"`
		InvoiceID       int64      `json:"invoice_id"`
		InvoiceNumber   string     `json:"invoice_number"`
		InvoiceCompany  string     `json:"invoice_company"`
		InvoiceDate     *time.Time `json:"invoice_date"`
		InvoiceStatus   string     `json:"invoice_status"`
	}{
		WorkorderID:     l.WorkorderID,
		WorkorderNumber: l.WorkorderNumber,
		Date:            l.Date,
		ProjectName:     l.ProjectName,
		FieldName:       l.FieldName,
		CropName:        l.CropName,
		LaborName:       l.LaborName,
		Contractor:      l.Contractor,
		SurfaceHa:       l.SurfaceHa.StringFixed(3),
		CostHa:          l.CostHa.StringFixed(3),
		CategoryName:    l.CategoryName,
		InvestorName:    l.InvestorName,
		USDAvgValue:     l.USDAvgValue.StringFixed(3),
		NetTotal:        l.NetTotal.StringFixed(0),    // Total $ Neto: entero más próximo
		TotalIVA:        l.TotalIVA.StringFixed(0),    // Total $ IVA: entero más próximo
		USDCostHa:       l.USDCostHa.StringFixed(3),
		USDNetTotal:     l.USDNetTotal.StringFixed(3),
		InvoiceID:       l.InvoiceID,
		InvoiceNumber:   l.InvoiceNumber,
		InvoiceCompany:  l.InvoiceCompany,
		InvoiceDate:     l.InvoiceDate,
		InvoiceStatus:   l.InvoiceStatus,
	}
	return json.Marshal(aux)
}

type ListLaborGroupResponse struct {
	PageInfo types.PageInfo  `json:"page_info"`
	Data     []LaborListItem `json:"data"`
}

func FromDomainListGroup(d *domain.LaborListItem) *LaborListItem {
	return &LaborListItem{
		WorkorderID:     d.WorkorderID,
		WorkorderNumber: d.WorkorderNumber,
		Date:            d.Date,
		ProjectName:     d.ProjectName,
		FieldName:       d.FieldName,
		CropName:        d.CropName,
		LaborName:       d.LaborName,
		Contractor:      d.Contractor,
		SurfaceHa:       d.SurfaceHa.Round(3),
		CostHa:          d.CostHa.Round(3),
		CategoryName:    d.CategoryName,
		InvestorName:    d.InvestorName,
		USDAvgValue:     d.USDAvgValue.Round(3),
		NetTotal:        d.NetTotal.Round(3),
		TotalIVA:        d.TotalIVA.Round(3),
		USDCostHa:       d.USDCostHa.Round(3),
		USDNetTotal:     d.USDNetTotal.Round(3),
		InvoiceID:       d.InvoiceID,
		InvoiceNumber:   d.InvoiceNumber,
		InvoiceCompany:  d.InvoiceCompany,
		InvoiceDate:     d.InvoiceDate,
		InvoiceStatus:   d.InvoiceStatus,
	}
}

func FromDomainList(pageInfo types.PageInfo, list []domain.LaborListItem) ListLaborGroupResponse {
	items := make([]LaborListItem, len(list))
	for i, d := range list {
		items[i] = *FromDomainListGroup(&d)
	}

	return ListLaborGroupResponse{PageInfo: pageInfo, Data: items}
}
