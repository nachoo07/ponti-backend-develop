package excel

var ColumnWidths = []float64{
	15, // INGRESO
	18, // N° REMITO
	15, // FECHA
	22, // INVERSOR
	25, // INSUMO
	15, // CANTIDAD
	15, // UNIDAD
	20, // RUBRO
	20, // TIPO/CLASE
	25, // PROVEEDOR
	15, // PRECIO U$
	18, // TOTAL U$
}

const (
	SheetName       = "Insumos"
	DateFormat      = "02/01/2006"
	DefaultFilename = "insumos.xlsx"
)
