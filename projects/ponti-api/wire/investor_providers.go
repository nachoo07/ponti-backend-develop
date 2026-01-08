package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
)

// ProvideInvestorRepository crea la implementación concreta de investor.Repository.
func ProvideInvestorRepository(repo investor.GormEnginePort) *investor.Repository {
	return investor.NewRepository(repo)
}

// ProvideInvestorRepositoryPort adapta *investor.Repository a la interfaz investor.RepositoryPort.
func ProvideInvestorRepositoryPort(r *investor.Repository) investor.RepositoryPort {
	return r
}

// ProvideInvestorUseCases agrupa repositorio en investor.UseCases.
func ProvideInvestorUseCases(
	rep investor.RepositoryPort,
) *investor.UseCases {
	return investor.NewUseCases(rep)
}

// ProvideInvestorUseCasesPort adapta *investor.UseCases a la interfaz investor.UseCasesPort.
func ProvideInvestorUseCasesPort(uc *investor.UseCases) investor.UseCasesPort {
	return uc
}

// ProvideInvestorHandler construye el handler HTTP para Investor.
func ProvideInvestorHandler(
	server investor.GinEnginePort,
	useCases investor.UseCasesPort,
	cfg investor.ConfigAPIPort,
	middlewares investor.MiddlewaresEnginePort,
) *investor.Handler {
	return investor.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideInvestorConfigAPI extrae la configuración específica de API para Investor.
func ProvideInvestorConfigAPI(cfg *config.Config) investor.ConfigAPIPort {
	return &cfg.API
}

// ProvideInvestorGormEnginePort adapta *pgorm.Repository a investor.GormEnginePort.
func ProvideInvestorGormEnginePort(r *pgorm.Repository) investor.GormEnginePort {
	return r
}

// ProvideInvestorGinEnginePort adapta *pgin.Server a investor.GinEnginePort.
func ProvideInvestorGinEnginePort(s *pgin.Server) investor.GinEnginePort {
	return s
}

// ProvideInvestorMiddlewaresEnginePort adapta *mwr.Middlewares a investor.MiddlewaresEnginePort.
func ProvideInvestorMiddlewaresEnginePort(m *mwr.Middlewares) investor.MiddlewaresEnginePort {
	return m
}

// InvestorSet expone todos los providers necesarios para Investor.
var InvestorSet = wire.NewSet(
	ProvideInvestorRepository,
	ProvideInvestorRepositoryPort,
	ProvideInvestorUseCases,
	ProvideInvestorUseCasesPort,
	ProvideInvestorHandler,
	ProvideInvestorConfigAPI,
	ProvideInvestorGormEnginePort,
	ProvideInvestorGinEnginePort,
	ProvideInvestorMiddlewaresEnginePort,
)
