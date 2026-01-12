package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	app_parameters "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/app_parameters"
)

// ProvideAppParametersGormEnginePort ...
func ProvideAppParametersGormEnginePort(r *gormpkg.Repository) app_parameters.GormEnginePort {
	return r
}

// ProvideAppParametersRepository ...
func ProvideAppParametersRepository(r app_parameters.GormEnginePort) *app_parameters.Repository {
	return app_parameters.NewRepository(r)
}

// ProvideAppParametersRepositoryPort ...
func ProvideAppParametersRepositoryPort(repo *app_parameters.Repository) app_parameters.RepositoryPort {
	return repo
}

// ProvideAppParametersConfigAPI ...
func ProvideAppParametersConfigAPI(c *cfg.Config) app_parameters.ConfigAPIPort {
	return &c.API
}

// ProvideAppParametersGinEnginePort ...
func ProvideAppParametersGinEnginePort(s *pgin.Server) app_parameters.GinEnginePort {
	return s
}

// ProvideAppParametersMiddlewaresEnginePort ...
func ProvideAppParametersMiddlewaresEnginePort(m *mwr.Middlewares) app_parameters.MiddlewaresEnginePort {
	return m
}

// ProvideAppParametersUseCases ...
func ProvideAppParametersUseCases(rep app_parameters.RepositoryPort) *app_parameters.UseCases {
	return app_parameters.NewUseCases(rep)
}

// ProvideAppParametersUseCasesPort ...
func ProvideAppParametersUseCasesPort(u *app_parameters.UseCases) app_parameters.UseCasesPort {
	return u
}

// ProvideAppParametersHandler ...
func ProvideAppParametersHandler(
	server app_parameters.GinEnginePort,
	ucs app_parameters.UseCasesPort,
	cfg app_parameters.ConfigAPIPort,
	mws app_parameters.MiddlewaresEnginePort,
) *app_parameters.Handler {
	return app_parameters.NewHandler(ucs, server, cfg, mws)
}

// AppParametersSet ...
var AppParametersSet = wire.NewSet(
	ProvideAppParametersGormEnginePort,
	ProvideAppParametersRepository,
	ProvideAppParametersRepositoryPort,
	ProvideAppParametersConfigAPI,
	ProvideAppParametersGinEnginePort,
	ProvideAppParametersMiddlewaresEnginePort,

	ProvideAppParametersUseCases,
	ProvideAppParametersUseCasesPort,
	ProvideAppParametersHandler,
)
