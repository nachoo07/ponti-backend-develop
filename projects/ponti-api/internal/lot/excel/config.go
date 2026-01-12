package excel

var ColumnWidths = []float64{
	20, // PROYECTO
	18, // CAMPO
	18, // LOTES
	24, // CULTIVO ANTERIOR
	24, // CULTIVO ACTUAL
	18, // VARIEDAD
	15, // SUP. SIEMBRA
	14, // FECHA SIEMBRA
	15, // SUP. COSECHA
	14, // FECHA COSECHA
	15, // TONELADAS
	18, // RENDIMIENTO
	20, // INGRESO NETO/HA
	18, // COSTO U$/HA
	18, // ARRIENDO/HA
	20, // ADMIN. PROYECTO/HA
	20, // ACTIVO TOTAL/HA
	22, // RESULTADO OPERATIVO
}

const (
	SheetName       = "Lotes"
	DateFormat      = "02/01/2006"
	DefaultFilename = "lotes.xlsx"
)
