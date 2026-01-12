package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type ExcelTableDto struct {
	WorkorderNumber string    `excel:"OT N°"`
	Date            time.Time `excel:"FECHA"`
	ProjectName     string    `excel:"PROYECTO"`
	FieldName       string    `excel:"CAMPO"`
	CropName        string    `excel:"CULTIVO"`
	LaborName       string    `excel:"LABOR"`
	Contractor      string    `excel:"CONTRATISTA"`
	SurfaceHa       float64   `excel:"SUPÉRFICIE"`
	CostHa          float64   `excel:"COSTO $/HECTÁREA"`
	InvestorName    string    `excel:"INVERSOR"`
	USDAvgValue     float64   `excel:"U$ PROM"`
	NetTotal        float64   `excel:"TOTAL $ / NETO"`
	TotalIVA        float64   `excel:"TOTAL $ / IVA"`
	USDCostHa       float64   `excel:"COSTO U$ /HA"`
	USDNetTotal     float64   `excel:"TOTAL U$ NETO"`

	InvoiceNumber  string     `excel:"N° FACTURA"`
	InvoiceCompany string     `excel:"EMPRESA"`
	InvoiceDate    *time.Time `excel:"FECHA DE FACTURACIÓN"`
	InvoiceStatus  string     `excel:"ESTADO DE FACTURA"`
}

func BuildExcelTableDTO(items []domain.LaborListItem) []ExcelDto {
	out := make([]ExcelDto, 0, len(items))
	for _, it := range items {
		out = append(out, ExcelDto{
			WorkorderNumber: it.WorkorderNumber,
			Date:            it.Date,
			ProjectName:     it.ProjectName,
			FieldName:       it.FieldName,
			CropName:        it.CropName,
			LaborName:       it.LaborName,
			Contractor:      it.Contractor,
			SurfaceHa:       decToFloat(it.SurfaceHa, 2),
			CostHa:          decToFloat(it.CostHa, 2),
			InvestorName:    it.InvestorName,
			USDAvgValue:     decToFloat(it.USDAvgValue, 2),
			NetTotal:        decToFloat(it.NetTotal, 2),
			TotalIVA:        decToFloat(it.TotalIVA, 2),
			USDCostHa:       decToFloat(it.USDCostHa, 2),
			USDNetTotal:     decToFloat(it.USDNetTotal, 2),
			InvoiceNumber:   it.InvoiceNumber,
			InvoiceCompany:  it.InvoiceCompany,
			InvoiceDate:     it.InvoiceDate,
			InvoiceStatus:   it.InvoiceStatus,
		})
	}
	return out
}
