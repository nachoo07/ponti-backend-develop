package wire

import (
	"os"
	"path/filepath"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgexcel "github.com/alphacodinggroup/ponti-backend/pkg/files-io/excel/excelize"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement"
	supplyExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/excel"
	"github.com/google/wire"
)

func ProvideSupplyMovementRepository(repo supply_movement.GormEnginePort) *supply_movement.Repository {
	return supply_movement.NewRepository(repo)
}

func ProvideSupplyMovementRepositoryPort(r *supply_movement.Repository) supply_movement.RepositoryPort {
	return r
}

type SupplyMovementExcelService struct {
	*pkgexcel.Service
}

// Crea el engine de Excel ya configurado
func ProvideSupplyMovementPkgExcelService() (*SupplyMovementExcelService, error) {
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
	return &SupplyMovementExcelService{s}, nil
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideSupplyMovementXLSXEnginePort(s *SupplyMovementExcelService) supply_movement.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideSupplyMovementExporterPort(eng supply_movement.XLSXEnginePort) supply_movement.ExporterAdapterPort {
	return supply_movement.NewExcelExporter(eng)
}

func ProvideSupplyMovementUseCases(suc *stock.UseCases, r *supply_movement.Repository, exp supply_movement.ExporterAdapterPort) *supply_movement.UseCases {
	return supply_movement.NewUseCases(r, suc, exp)
}

func ProvideSupplyMovementUseCasesPort(uc *supply_movement.UseCases) supply_movement.UseCasesPort {
	return uc
}

func ProvideSupplyMovementHandler(
	server supply_movement.GinEnginePort,
	useCases supply_movement.UseCasesPort,
	cfg supply_movement.ConfigAPIPort,
	middlewares supply_movement.MiddlewaresEnginePort,
	ucps project.UseCasesPort,
) *supply_movement.Handler {
	return supply_movement.NewHandler(useCases, server, cfg, middlewares, ucps)
}

func ProvideSupplyMovementConfigAPI(cfg *config.Config) supply_movement.ConfigAPIPort {
	return &cfg.API
}

func ProvideSupplyMovementGormEnginePort(r *pgorm.Repository) supply_movement.GormEnginePort {
	return r
}

func ProvideSupplyMovementGinEnginePort(s *pgin.Server) supply_movement.GinEnginePort {
	return s
}

func ProvideSupplyMovementMiddlewaresEnginePort(m *mwr.Middlewares) supply_movement.MiddlewaresEnginePort {
	return m
}

var SupplyMovementSet = wire.NewSet(
	ProvideSupplyMovementRepository,
	ProvideSupplyMovementRepositoryPort,
	ProvideSupplyMovementUseCases,
	ProvideSupplyMovementUseCasesPort,
	ProvideSupplyMovementHandler,
	ProvideSupplyMovementConfigAPI,
	ProvideSupplyMovementGormEnginePort,
	ProvideSupplyMovementGinEnginePort,
	ProvideSupplyMovementMiddlewaresEnginePort,
	ProvideSupplyMovementPkgExcelService,
	ProvideSupplyMovementExporterPort,
	ProvideSupplyMovementXLSXEnginePort,
)
