package wire

import (
	"os"
	"path/filepath"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgexcel "github.com/alphacodinggroup/ponti-backend/pkg/files-io/excel/excelize"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor"
	labexcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/excel"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/google/wire"
)

// ProvideLaborRepository crea la implementación concreta de labor.Repository.
func ProvideLaborRepository(repo labor.GormEnginePort) *labor.Repository {
	return labor.NewRepository(repo)
}

func ProvideLaborRepositoryPort(r *labor.Repository) labor.RepositoryPort {
	return r
}

type LaborExcelService struct {
	*pkgexcel.Service
}

// Crea el engine de Excel ya configurado
func ProvideLaborPkgExcelService() (*LaborExcelService, error) {
	fp := filepath.Join(os.TempDir(), labexcel.DefaultFilename)
	write := true
	s, err := pkgexcel.Bootstrap(fp,
		labexcel.SheetName,
		labexcel.DateFormat,
		&write,
		labexcel.ColumnWidths,
	)
	if err != nil {
		return nil, err
	}
	return &LaborExcelService{s}, nil
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideLaborXLSXEnginePort(s *LaborExcelService) labor.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideLaborExporterPort(eng labor.XLSXEnginePort) labor.ExporterAdapterPort {
	return labor.NewExcelExporter(eng)
}

// ProvideLaborUseCases agrupa repositorio y servicio
func ProvideLaborUseCases(rep labor.RepositoryPort, exp labor.ExporterAdapterPort) *labor.UseCases {
	return labor.NewUseCases(rep, exp)
}

func ProvideLaborUseCasesPort(uc *labor.UseCases) labor.UseCasesPort {
	return uc
}

func ProvideLaborHandler(
	server labor.GinEnginePort,
	useCases labor.UseCasesPort,
	cfg labor.ConfigAPIPort,
	middlewares labor.MiddlewaresEnginePort,
	useCaseProject project.UseCasesPort) *labor.Handler {
	return labor.NewHandler(useCases, server, cfg, middlewares, useCaseProject)
}

func ProvideLaborConfigAPI(cfg *config.Config) labor.ConfigAPIPort {
	return &cfg.API
}

func ProvideLaborGormEnginePort(r *pgorm.Repository) labor.GormEnginePort {
	return r
}

func ProvideLaborGinEnginePort(s *pgin.Server) labor.GinEnginePort {
	return s
}

func ProvideLaborMiddlewaresEnginePort(m *mwr.Middlewares) labor.MiddlewaresEnginePort {
	return m
}

var LaborSet = wire.NewSet(
	ProvideLaborRepository,
	ProvideLaborRepositoryPort,
	ProvideLaborUseCases,
	ProvideLaborUseCasesPort,
	ProvideLaborHandler,
	ProvideLaborConfigAPI,
	ProvideLaborGormEnginePort,
	ProvideLaborGinEnginePort,
	ProvideLaborMiddlewaresEnginePort,
	ProvideLaborPkgExcelService,
	ProvideLaborXLSXEnginePort,
	ProvideLaborExporterPort,
)
