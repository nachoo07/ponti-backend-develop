package excel

var ColumnWidths = []float64{
	25, // Insumo
	18, // Rubro
	22, // Inversor
	14, // Ingresados
	14, // Consumidos
	14, // Stock
	14, // Stock Real
	14, // Diferencia
	16, // Fecha de Cierre
	14, // Precio U.
	16, // Total U$
}

const (
	SheetName       = "Stock"
	DateFormat      = "02/01/2006"
	DefaultFilename = "stock.xlsx"
)
