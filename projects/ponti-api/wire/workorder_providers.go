package wire

import (
	"os"
	"path/filepath"

	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgexcel "github.com/alphacodinggroup/ponti-backend/pkg/files-io/excel/excelize"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	workorder "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder"
	workorderexcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/excel"
)

// ProvideWorkorderRepository crea la implementación concreta de workorder.Repository.
func ProvideWorkorderRepository(repo workorder.GormEngine) *workorder.Repository {
	return workorder.NewRepository(repo)
}

// ProvideWorkorderRepositoryPort adapta *workorder.Repository a la interfaz workorder.RepositoryPort.
func ProvideWorkorderRepositoryPort(r *workorder.Repository) workorder.RepositoryPort {
	return r
}

// Crea el engine de Excel ya configurado
func ProvidePkgExcelService() (*pkgexcel.Service, error) {
	fp := filepath.Join(os.TempDir(), workorderexcel.DefaultFilename)
	write := true
	return pkgexcel.Bootstrap(
		fp,
		workorderexcel.SheetName,
		workorderexcel.DateFormat,
		&write,
		workorderexcel.ColumnWidths,
	)
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideXLSXEnginePort(s *pkgexcel.Service) workorder.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideExporterPort(eng workorder.XLSXEnginePort) workorder.ExporterAdapterPort {
	return workorder.NewExcelExporter(eng)
}

// ProvideWorkorderUseCases agrupa repositorios en workorder.UseCases.
func ProvideWorkorderUseCases(repo workorder.RepositoryPort, exp workorder.ExporterAdapterPort) *workorder.UseCases {
	return workorder.NewUseCases(repo, exp)
}

// ProvideWorkorderUseCasesPort adapta *workorder.UseCases a la interfaz workorder.UseCasesPort.
func ProvideWorkorderUseCasesPort(uc *workorder.UseCases) workorder.UseCasesPort {
	return uc
}

// ProvideWorkorderHandler construye el handler HTTP para Workorder.
func ProvideWorkorderHandler(
	server workorder.GinEnginePort,
	useCases workorder.UseCasesPort,
	cfg workorder.ConfigAPIPort,
	middlewares workorder.MiddlewaresEnginePort,
) *workorder.Handler {
	return workorder.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideWorkorderConfigAPI extrae la configuración específica de API para Workorder.
func ProvideWorkorderConfigAPI(cfg *config.Config) workorder.ConfigAPIPort {
	return &cfg.API
}

// ProvideWorkorderGormEnginePort adapta *pgorm.Repository a workorder.GormEngine.
func ProvideWorkorderGormEnginePort(r *pgorm.Repository) workorder.GormEngine {
	return r
}

// ProvideWorkorderGinEnginePort adapta *pgin.Server a workorder.GinEnginePort.
func ProvideWorkorderGinEnginePort(s *pgin.Server) workorder.GinEnginePort {
	return s
}

// ProvideWorkorderMiddlewaresEnginePort adapta *mwr.Middlewares a workorder.MiddlewaresEnginePort.
func ProvideWorkorderMiddlewaresEnginePort(m *mwr.Middlewares) workorder.MiddlewaresEnginePort {
	return m
}

// WorkorderSet expone todos los providers necesarios para Workorder.
var WorkorderSet = wire.NewSet(
	ProvideWorkorderRepository,
	ProvideWorkorderRepositoryPort,
	ProvideWorkorderUseCases,
	ProvideWorkorderUseCasesPort,
	ProvideWorkorderHandler,
	ProvideWorkorderConfigAPI,
	ProvideWorkorderGormEnginePort,
	ProvideWorkorderGinEnginePort,
	ProvideWorkorderMiddlewaresEnginePort,
	ProvidePkgExcelService,
	ProvideExporterPort,
	ProvideXLSXEnginePort,
)
