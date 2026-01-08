package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
)

// ProvideManagerRepository crea la implementación concreta de manager.Repository.
func ProvideManagerRepository(repo manager.GormEnginePort) *manager.Repository {
	return manager.NewRepository(repo)
}

// ProvideManagerRepositoryPort adapta *manager.Repository a la interfaz manager.RepositoryPort.
func ProvideManagerRepositoryPort(r *manager.Repository) manager.RepositoryPort {
	return r
}

// ProvideManagerUseCases agrupa repositorio en manager.UseCases.
func ProvideManagerUseCases(
	rep manager.RepositoryPort,
) *manager.UseCases {
	return manager.NewUseCases(rep)
}

// ProvideManagerUseCasesPort adapta *manager.UseCases a la interfaz manager.UseCasesPort.
func ProvideManagerUseCasesPort(uc *manager.UseCases) manager.UseCasesPort {
	return uc
}

// ProvideManagerHandler construye el handler HTTP para Manager.
func ProvideManagerHandler(
	server manager.GinEnginePort,
	useCases manager.UseCasesPort,
	cfg manager.ConfigAPIPort,
	middlewares manager.MiddlewaresEnginePort,
) *manager.Handler {
	return manager.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideManagerConfigAPI extrae la configuración específica de API para Manager.
func ProvideManagerConfigAPI(cfg *config.Config) manager.ConfigAPIPort {
	return &cfg.API
}

// ProvideManagerGormEnginePort adapta *pgorm.Repository a manager.GormEnginePort.
func ProvideManagerGormEnginePort(r *pgorm.Repository) manager.GormEnginePort {
	return r
}

// ProvideManagerGinEnginePort adapta *pgin.Server a manager.GinEnginePort.
func ProvideManagerGinEnginePort(s *pgin.Server) manager.GinEnginePort {
	return s
}

// ProvideManagerMiddlewaresEnginePort adapta *mwr.Middlewares a manager.MiddlewaresEnginePort.
func ProvideManagerMiddlewaresEnginePort(m *mwr.Middlewares) manager.MiddlewaresEnginePort {
	return m
}

// ManagerSet expone todos los providers necesarios para Manager.
var ManagerSet = wire.NewSet(
	ProvideManagerRepository,
	ProvideManagerRepositoryPort,
	ProvideManagerUseCases,
	ProvideManagerUseCasesPort,
	ProvideManagerHandler,
	ProvideManagerConfigAPI,
	ProvideManagerGormEnginePort,
	ProvideManagerGinEnginePort,
	ProvideManagerMiddlewaresEnginePort,
)
