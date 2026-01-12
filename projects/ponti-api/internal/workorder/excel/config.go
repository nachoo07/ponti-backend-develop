package excel

var ColumnWidths = []float64{
	15, // NUMERO DE ORDEN
	20, // PROYECTO
	18, // CAMPO
	15, // LOTE
	15, // FECHA
	18, // CULTIVO
	20, // LABOR
	20, // TIPO/CLASE
	25, // CONTRATISTA
	15, // SUPERFICIE
	22, // INSUMO
	15, // CONSUMO
	18, // RUBRO
	15, // DOSIS
	18, // COST U$/HA
	18, // PRECIO UNIDAD
	18, // TOTAL COSTO
}

const (
	SheetName       = "Órdenes de trabajos"
	DateFormat      = "02/01/2006"
	DefaultFilename = "ordenes_de_trabajos.xlsx"
)
