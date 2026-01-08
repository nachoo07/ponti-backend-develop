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
	supply "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply"
	supplyExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/excel"
)

// ProvideSupplyRepository crea la implementación concreta de supply.Repository.
func ProvideSupplyRepository(repo supply.GormEnginePort) *supply.Repository {
	return supply.NewRepository(repo)
}

// ProvideSupplyRepositoryPort adapta *supply.Repository a la interfaz supply.RepositoryPort.
func ProvideSupplyRepositoryPort(r *supply.Repository) supply.RepositoryPort {
	return r
}

type SupplyExcelService struct {
	*pkgexcel.Service
}

// Crea el engine de Excel ya configurado
func ProvideSupplyPkgExcelService() (*SupplyExcelService, error) {
	fp := filepath.Join(os.TempDir(), supplyExcel.DefaultFilename)
	write := true
	s, err := pkgexcel.Bootstrap(fp,
		supplyExcel.SheetName,
		supplyExcel.DateFormat,
		&write,
		supplyExcel.ColumnWidths,
	)
	if err != nil {
		return nil, err
	}
	return &SupplyExcelService{s}, nil
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideSupplyXLSXEnginePort(s *SupplyMovementExcelService) supply.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideSupplyExporterPort(eng supply.XLSXEnginePort) supply.ExporterAdapterPort {
	return supply.NewExcelExporter(eng)
}

// ProvideSupplyUseCases agrupa repositorios en supply.UseCases.
func ProvideSupplyUseCases(repo supply.RepositoryPort, exp supply.ExporterAdapterPort) *supply.UseCases {
	return supply.NewUseCases(repo, exp)
}

// ProvideSupplyUseCasesPort adapta *supply.UseCases a la interfaz supply.UseCasesPort.
func ProvideSupplyUseCasesPort(uc *supply.UseCases) supply.UseCasesPort {
	return uc
}

// ProvideSupplyHandler construye el handler HTTP para Supply.
func ProvideSupplyHandler(
	server supply.GinEnginePort,
	useCases supply.UseCasesPort,
	cfg supply.ConfigAPIPort,
	middlewares supply.MiddlewaresEnginePort,
) *supply.Handler {
	return supply.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideSupplyConfigAPI extrae la configuración específica de API para Supply.
func ProvideSupplyConfigAPI(cfg *config.Config) supply.ConfigAPIPort {
	return &cfg.API
}

// ProvideSupplyGormEnginePort adapta *pgorm.Repository a supply.GormEnginePort.
func ProvideSupplyGormEnginePort(r *pgorm.Repository) supply.GormEnginePort {
	return r
}

// ProvideSupplyGinEnginePort adapta *pgin.Server a supply.GinEnginePort.
func ProvideSupplyGinEnginePort(s *pgin.Server) supply.GinEnginePort {
	return s
}

// ProvideSupplyMiddlewaresEnginePort adapta *mwr.Middlewares a supply.MiddlewaresEnginePort.
func ProvideSupplyMiddlewaresEnginePort(m *mwr.Middlewares) supply.MiddlewaresEnginePort {
	return m
}

// SupplySet expone todos los providers necesarios para Supply.
var SupplySet = wire.NewSet(
	ProvideSupplyRepository,
	ProvideSupplyRepositoryPort,
	ProvideSupplyUseCases,
	ProvideSupplyUseCasesPort,
	ProvideSupplyHandler,
	ProvideSupplyConfigAPI,
	ProvideSupplyGormEnginePort,
	ProvideSupplyGinEnginePort,
	ProvideSupplyMiddlewaresEnginePort,
	ProvideSupplyPkgExcelService,
	ProvideSupplyExporterPort,
	ProvideSupplyXLSXEnginePort,
)
