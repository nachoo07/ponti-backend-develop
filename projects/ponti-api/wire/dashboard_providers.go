package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	dashboard "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard"
)

// ProvideDashboardRepository crea la implementación concreta de dashboard.Repository.
func ProvideDashboardRepository(repo dashboard.GormEnginePort) *dashboard.Repository {
	return dashboard.NewRepository(repo)
}

// ProvideDashboardRepositoryPort adapta *dashboard.Repository a la interfaz dashboard.RepositoryPort.
func ProvideDashboardRepositoryPort(r *dashboard.Repository) dashboard.RepositoryPort {
	return r
}

// ProvideDashboardUseCases agrupa repositorio en dashboard.UseCases.
func ProvideDashboardUseCases(
	rep dashboard.RepositoryPort,
) *dashboard.UseCases {
	return dashboard.NewUseCases(rep)
}

// ProvideDashboardUseCasesPort adapta *dashboard.UseCases a la interfaz dashboard.UseCasesPort.
func ProvideDashboardUseCasesPort(uc *dashboard.UseCases) dashboard.UseCasesPort {
	return uc
}

// ProvideDashboardHandler construye el handler HTTP para Dashboard.
func ProvideDashboardHandler(
	server dashboard.GinEnginePort,
	useCases dashboard.UseCasesPort,
	cfg dashboard.ConfigAPIPort,
	middlewares dashboard.MiddlewaresEnginePort,
) *dashboard.Handler {
	return dashboard.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideDashboardConfigAPI extrae la configuración específica de API para Dashboard.
func ProvideDashboardConfigAPI(cfg *config.Config) dashboard.ConfigAPIPort {
	return &cfg.API
}

// ProvideDashboardGormEnginePort adapta *pgorm.Repository a dashboard.GormEnginePort.
func ProvideDashboardGormEnginePort(r *pgorm.Repository) dashboard.GormEnginePort {
	return r
}

// ProvideDashboardGinEnginePort adapta *pgin.Server a dashboard.GinEnginePort.
func ProvideDashboardGinEnginePort(s *pgin.Server) dashboard.GinEnginePort {
	return s
}

// ProvideDashboardMiddlewaresEnginePort adapta *mwr.Middlewares a dashboard.MiddlewaresEnginePort.
func ProvideDashboardMiddlewaresEnginePort(m *mwr.Middlewares) dashboard.MiddlewaresEnginePort {
	return m
}

// DashboardSet expone todos los providers necesarios para Dashboard.
var DashboardSet = wire.NewSet(
	ProvideDashboardRepository,
	ProvideDashboardRepositoryPort,
	ProvideDashboardUseCases,
	ProvideDashboardUseCasesPort,
	ProvideDashboardHandler,
	ProvideDashboardConfigAPI,
	ProvideDashboardGormEnginePort,
	ProvideDashboardGinEnginePort,
	ProvideDashboardMiddlewaresEnginePort,
)
