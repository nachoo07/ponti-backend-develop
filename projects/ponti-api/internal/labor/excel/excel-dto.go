package excel

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/shopspring/decimal"
)

type ExcelDto struct {
	WorkorderNumber string    `excel:"OT N°"`
	Date            time.Time `excel:"FECHA"`
	ProjectName     string    `excel:"PROYECTO"`
	FieldName       string    `excel:"CAMPO"`
	CropName        string    `excel:"CULTIVO"`
	LaborName       string    `excel:"LABOR"`
	Contractor      string    `excel:"CONTRATISTA"`
	SurfaceHa       string    `excel:"SUPÉRFICIE"`
	CostHa          string    `excel:"COSTO $/HECTÁREA"`
	InvestorName    string    `excel:"INVERSOR"`
	USDAvgValue     float64   `excel:"U$ PROM"`
	NetTotal        string    `excel:"TOTAL $ / NETO"`
	TotalIVA        string    `excel:"TOTAL $ / IVA"`
	USDCostHa       float64   `excel:"COSTO U$ /HA"`
	USDNetTotal     float64   `excel:"TOTAL U$ NETO"`

	InvoiceNumber  string     `excel:"N° FACTURA"`
	InvoiceCompany string     `excel:"EMPRESA"`
	InvoiceDate    *time.Time `excel:"FECHA DE FACTURACIÓN"`
	InvoiceStatus  string     `excel:"ESTADO DE FACTURA"`
}

func BuildExcelDTO(items []domain.LaborListItem) []ExcelDto {
	out := make([]ExcelDto, 0, len(items))

	for _, it := range items {
		// InvoiceDate ya es *time.Time, no necesita conversión
		invDate := it.InvoiceDate

		out = append(out, ExcelDto{
			WorkorderNumber: it.WorkorderNumber,
			Date:            it.Date,
			ProjectName:     it.ProjectName,
			FieldName:       it.FieldName,
			CropName:        it.CropName,
			LaborName:       it.LaborName,
			Contractor:      it.Contractor,
			SurfaceHa:       decToString(it.SurfaceHa, 2),
			CostHa:          decToString(it.CostHa, 2),
			InvestorName:    it.InvestorName,
			USDAvgValue:     decToFloat(it.USDAvgValue, 2),
			NetTotal:        decToString(it.NetTotal, 2),
			TotalIVA:        decToString(it.TotalIVA, 2),
			USDCostHa:       decToFloat(it.USDCostHa, 2),
			USDNetTotal:     decToFloat(it.USDNetTotal, 2),
			InvoiceNumber:   it.InvoiceNumber,
			InvoiceCompany:  it.InvoiceCompany,
			InvoiceDate:     invDate,
			InvoiceStatus:   it.InvoiceStatus,
		})
	}

	return out
}

// helper para convertir decimal en float64 con redondeo opcional
func decToFloat(d decimal.Decimal, scale int32) float64 {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return f
}

func decToString(d decimal.Decimal, scale int32) string {
	if scale < 0 {
		return d.String()
	}
	return d.Round(scale).StringFixed(scale)
}
