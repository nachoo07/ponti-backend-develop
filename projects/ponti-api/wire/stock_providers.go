package wire

import (
	"os"
	"path/filepath"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgexcel "github.com/alphacodinggroup/ponti-backend/pkg/files-io/excel/excelize"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	stockExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/excel"

	"github.com/google/wire"
)

// ProvideStockRepository creates the concrete implementation of stock.Repository.
func ProvideStockRepository(repo stock.GormEnginePort) *stock.Repository {
	return stock.NewRepository(repo)
}

func ProvideStockRepositoryPort(r *stock.Repository) stock.RepositoryPort {
	return r
}

type StockExcelService struct {
	*pkgexcel.Service
}

// Crea el engine de Excel ya configurado
func ProvideStockPkgExcelService() (*StockExcelService, error) {
	fp := filepath.Join(os.TempDir(), stockExcel.DefaultFilename)
	write := true
	s, err := pkgexcel.Bootstrap(fp,
		stockExcel.SheetName,
		stockExcel.DateFormat,
		&write,
		stockExcel.ColumnWidths,
	)
	if err != nil {
		return nil, err
	}
	return &StockExcelService{s}, nil
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideStockXLSXEnginePort(s *SupplyMovementExcelService) stock.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideStockExporterPort(eng stock.XLSXEnginePort) stock.ExporterAdapterPort {
	return stock.NewExcelExporter(eng)
}

// ProvideStockUseCases groups repository into stock.UseCases.
func ProvideStockUseCases(rep stock.RepositoryPort, excel stock.ExporterAdapterPort) *stock.UseCases {
	return stock.NewUseCases(rep, excel)
}

func ProvideStockUseCasesPort(uc *stock.UseCases) stock.UseCasesPort {
	return uc
}

func ProvideStockHandler(
	server stock.GinEnginePort,
	useCases stock.UseCasesPort,
	cfg stock.ConfigAPIPort,
	middlewares stock.MiddlewaresEnginePort,
	ucps project.UseCasesPort,
) *stock.Handler {
	return stock.NewHandler(useCases, server, cfg, middlewares, ucps)
}

func ProvideStockConfigAPI(cfg *config.Config) stock.ConfigAPIPort {
	return &cfg.API
}

func ProvideStockGormEnginePort(r *pgorm.Repository) stock.GormEnginePort {
	return r
}

func ProvideStockGinEnginePort(s *pgin.Server) stock.GinEnginePort {
	return s
}

func ProvideStockMiddlewaresEnginePort(m *mwr.Middlewares) stock.MiddlewaresEnginePort {
	return m
}

var StockSet = wire.NewSet(
	ProvideStockRepository,
	ProvideStockRepositoryPort,
	ProvideStockUseCases,
	ProvideStockUseCasesPort,
	ProvideStockHandler,
	ProvideStockConfigAPI,
	ProvideStockGormEnginePort,
	ProvideStockGinEnginePort,
	ProvideStockMiddlewaresEnginePort,
	ProvideStockPkgExcelService,
	ProvideStockExporterPort,
	ProvideStockXLSXEnginePort,
)
