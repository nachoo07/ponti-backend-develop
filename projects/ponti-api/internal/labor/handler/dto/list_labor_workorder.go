package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type LaborRawItem struct {
	WorkorderNumber string          `json:"workorder_number"`
	Date            time.Time       `json:"date"`
	ProjectName     string          `json:"project_name"`
	FieldName       string          `json:"field_name"`
	CropName        string          `json:"crop_name"`
	LaborName       string          `json:"labor_name"`
	Contractor      string          `json:"contractor"`
	SurfaceHa       decimal.Decimal `json:"effective_area"`
	CostHa          decimal.Decimal `json:"price"`
	CategoryName    string          `json:"contractor_name"`
	InvestorName    string          `json:"investor_name"`
	USDAvgValue     decimal.Decimal `json:"usd_avg_value"`
	InvoiceNumber   string          `json:"invoice_number"`
	InvoiceCompany  string          `json:"invoice_company"`
	InvoiceDate     *time.Time      `json:"invoice_date"`
	InvoiceStatus   string          `json:"invoice_status"`
}

type LaborByWorkorderListResponse struct {
	Data []LaborRawItem `json:"data"`
}

func ToLaborListResponse(items []domain.LaborRawItem) LaborByWorkorderListResponse {
	dtos := make([]LaborRawItem, len(items))
	for i, d := range items {
		dtos[i] = LaborRawItem{
			WorkorderNumber: d.WorkorderNumber,
			Date:            d.Date,
			ProjectName:     d.ProjectName,
			FieldName:       d.FieldName,
			CropName:        d.CropName,
			LaborName:       d.LaborName,
			Contractor:      d.Contractor,
			SurfaceHa:       d.SurfaceHa,
			CostHa:          d.CostHa,
			CategoryName:    d.CategoryName,
			InvestorName:    d.InvestorName,
			USDAvgValue:     d.USDAvgValue,
			InvoiceNumber:   d.InvoiceNumber,
			InvoiceCompany:  d.InvoiceCompany,
			InvoiceDate:     d.InvoiceDate,
			InvoiceStatus:   d.InvoiceStatus,
		}
	}
	return LaborByWorkorderListResponse{Data: dtos}
}
