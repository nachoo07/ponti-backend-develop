package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	dashboard "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard"
	data_integrity "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/data-integrity"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	report "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report"
	stock "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	workorder "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder"
)

// ProvideDataIntegrityUseCases construye los casos de uso de data_integrity
func ProvideDataIntegrityUseCases(
	workorderRepo data_integrity.WorkorderRepositoryPort,
	dashboardRepo data_integrity.DashboardRepositoryPort,
	lotRepo data_integrity.LotRepositoryPort,
	reportRepo data_integrity.ReportRepositoryPort,
	stockRepo data_integrity.StockRepositoryPort,
) *data_integrity.UseCases {
	return data_integrity.NewUseCases(
		workorderRepo,
		dashboardRepo,
		lotRepo,
		reportRepo,
		stockRepo,
	)
}

// ProvideDataIntegrityUseCasesPort adapta *data_integrity.UseCases a la interfaz data_integrity.UseCasesPort
func ProvideDataIntegrityUseCasesPort(uc *data_integrity.UseCases) data_integrity.UseCasesPort {
	return uc
}

// ProvideDataIntegrityHandler construye el handler HTTP para Data Integrity
func ProvideDataIntegrityHandler(
	server data_integrity.GinEnginePort,
	useCases data_integrity.UseCasesPort,
	cfg data_integrity.ConfigAPIPort,
	middlewares data_integrity.MiddlewaresEnginePort,
) *data_integrity.Handler {
	return data_integrity.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideDataIntegrityConfigAPI extrae la configuración específica de API para Data Integrity
func ProvideDataIntegrityConfigAPI(cfg *config.Config) data_integrity.ConfigAPIPort {
	return &cfg.API
}

// ProvideDataIntegrityGinEnginePort adapta *pgin.Server a data_integrity.GinEnginePort
func ProvideDataIntegrityGinEnginePort(s *pgin.Server) data_integrity.GinEnginePort {
	return s
}

// ProvideDataIntegrityMiddlewaresEnginePort adapta *mwr.Middlewares a data_integrity.MiddlewaresEnginePort
func ProvideDataIntegrityMiddlewaresEnginePort(m *mwr.Middlewares) data_integrity.MiddlewaresEnginePort {
	return m
}

// ProvideDataIntegrityWorkorderRepositoryPort adapta workorder.RepositoryPort a data_integrity.WorkorderRepositoryPort
func ProvideDataIntegrityWorkorderRepositoryPort(r workorder.RepositoryPort) data_integrity.WorkorderRepositoryPort {
	return r
}

// ProvideDataIntegrityDashboardRepositoryPort adapta dashboard.RepositoryPort a data_integrity.DashboardRepositoryPort
func ProvideDataIntegrityDashboardRepositoryPort(r dashboard.RepositoryPort) data_integrity.DashboardRepositoryPort {
	return r
}

// ProvideDataIntegrityLotRepositoryPort adapta lot.RepositoryPort a data_integrity.LotRepositoryPort
func ProvideDataIntegrityLotRepositoryPort(r lot.RepositoryPort) data_integrity.LotRepositoryPort {
	return r
}

// ProvideDataIntegrityReportRepositoryPort adapta report.ReportRepositoryPort a data_integrity.ReportRepositoryPort
func ProvideDataIntegrityReportRepositoryPort(r report.ReportRepositoryPort) data_integrity.ReportRepositoryPort {
	return r
}

// ProvideDataIntegrityStockRepositoryPort adapta stock.RepositoryPort a data_integrity.StockRepositoryPort
func ProvideDataIntegrityStockRepositoryPort(r stock.RepositoryPort) data_integrity.StockRepositoryPort {
	return r
}

// DataIntegritySet expone todos los providers necesarios para Data Integrity
var DataIntegritySet = wire.NewSet(
	ProvideDataIntegrityUseCases,
	ProvideDataIntegrityUseCasesPort,
	ProvideDataIntegrityHandler,
	ProvideDataIntegrityConfigAPI,
	ProvideDataIntegrityGinEnginePort,
	ProvideDataIntegrityMiddlewaresEnginePort,
	ProvideDataIntegrityWorkorderRepositoryPort,
	ProvideDataIntegrityDashboardRepositoryPort,
	ProvideDataIntegrityLotRepositoryPort,
	ProvideDataIntegrityReportRepositoryPort,
	ProvideDataIntegrityStockRepositoryPort,
)
