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

	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	lotExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/excel"
)

// ProvideLotRepository crea la implementación concreta de lot.Repository.
func ProvideLotRepository(repo lot.GormEnginePort) *lot.Repository {
	return lot.NewRepository(repo)
}

// ProvideLotRepositoryPort adapta *lot.Repository a la interfaz lot.RepositoryPort.
func ProvideLotRepositoryPort(r *lot.Repository) lot.RepositoryPort {
	return r
}

type LotExcelService struct {
	*pkgexcel.Service
}

// Crea el engine de Excel ya configurado
func ProvideLotPkgExcelService() (*LotExcelService, error) {
	fp := filepath.Join(os.TempDir(), lotExcel.DefaultFilename)
	write := true
	s, err := pkgexcel.Bootstrap(fp,
		lotExcel.SheetName,
		lotExcel.DateFormat,
		&write,
		lotExcel.ColumnWidths,
	)
	if err != nil {
		return nil, err
	}
	return &LotExcelService{s}, nil
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideLotXLSXEnginePort(s *LotExcelService) lot.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideLotExporterPort(eng lot.XLSXEnginePort) lot.ExporterAdapterPort {
	return lot.NewExcelExporter(eng)
}

// ProvideLotUseCases agrupa repositorio y servicio de crop en lot.UseCases.
func ProvideLotUseCases(
	rep lot.RepositoryPort,
	exp lot.ExporterAdapterPort,
) *lot.UseCases {
	return lot.NewUseCases(rep, exp)
}

// ProvideLotUseCasesPort adapta *lot.UseCases a la interfaz lot.UseCasesPort.
func ProvideLotUseCasesPort(uc *lot.UseCases) lot.UseCasesPort {
	return uc
}

// ProvideLotHandler construye el handler HTTP para Lot.
func ProvideLotHandler(
	server lot.GinEnginePort,
	useCases lot.UseCasesPort,
	cfg lot.ConfigAPIPort,
	middlewares lot.MiddlewaresEnginePort,
) *lot.Handler {
	return lot.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideLotConfigAPI extrae la configuración específica de API para Lot.
func ProvideLotConfigAPI(cfg *config.Config) lot.ConfigAPIPort {
	return &cfg.API
}

// ProvideLotGormEnginePort adapta *pgorm.Repository a lot.GormEnginePort.
func ProvideLotGormEnginePort(r *pgorm.Repository) lot.GormEnginePort {
	return r
}

// ProvideLotGinEnginePort adapta *pgin.Server a lot.GinEnginePort.
func ProvideLotGinEnginePort(s *pgin.Server) lot.GinEnginePort {
	return s
}

// ProvideLotMiddlewaresEnginePort adapta *mwr.Middlewares a lot.MiddlewaresEnginePort.
func ProvideLotMiddlewaresEnginePort(m *mwr.Middlewares) lot.MiddlewaresEnginePort {
	return m
}

// LotSet expone todos los providers necesarios para Lot.
var LotSet = wire.NewSet(
	ProvideLotRepository,
	ProvideLotRepositoryPort,
	ProvideLotUseCases,
	ProvideLotUseCasesPort,
	ProvideLotHandler,
	ProvideLotConfigAPI,
	ProvideLotGormEnginePort,
	ProvideLotGinEnginePort,
	ProvideLotMiddlewaresEnginePort,
	ProvideLotPkgExcelService,
	ProvideLotExporterPort,
	ProvideLotXLSXEnginePort,
)
