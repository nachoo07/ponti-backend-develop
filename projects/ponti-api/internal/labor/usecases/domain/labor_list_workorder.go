package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type LaborListItem struct {
	WorkorderID     int64           // id de la orden
	WorkorderNumber string          //numero de orden
	Date            time.Time       // fecha de la orden
	ProjectName     string          // nombre del proyecto
	FieldName       string          // nombre del campo
	CropName        string          // nombre del cultivo
	LaborName       string          // nombre de la labor
	Contractor      string          // nombre del contratista
	SurfaceHa       decimal.Decimal // superficie
	CostHa          decimal.Decimal // costo por ha
	CategoryName    string          // rubro
	InvestorName    string          // nombre del inversor
	USDAvgValue     decimal.Decimal // valor dolar promedio
	NetTotal        decimal.Decimal // total neto
	TotalIVA        decimal.Decimal // total IVA
	USDCostHa       decimal.Decimal // costo de dolar por ha
	USDNetTotal     decimal.Decimal // total neto dolar

	InvoiceID      int64      // id de la factura
	InvoiceNumber  string     // numero de factura
	InvoiceCompany string     // empresa de la factura
	InvoiceDate    *time.Time // fecha de factura
	InvoiceStatus  string     // estado de factura
}

type LaborRawItem struct {
	WorkorderID     int64
	WorkorderNumber string
	Date            time.Time
	ProjectName     string
	FieldName       string
	CropName        string
	LaborName       string
	Contractor      string
	SurfaceHa       decimal.Decimal
	CostHa          decimal.Decimal
	CategoryName    string
	InvestorName    string
	USDAvgValue     decimal.Decimal

	// Campos calculados de la vista fix_labors_list
	NetTotal    decimal.Decimal
	TotalIVA    decimal.Decimal
	USDCostHa   decimal.Decimal
	USDNetTotal decimal.Decimal

	InvoiceID      int64
	InvoiceNumber  string
	InvoiceCompany string
	InvoiceDate    *time.Time
	InvoiceStatus  string
}
