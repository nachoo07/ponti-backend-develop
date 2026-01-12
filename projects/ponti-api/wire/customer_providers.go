package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
)

// ProvideCustomerRepository crea la implementación concreta de customer.Repository.
func ProvideCustomerRepository(repo customer.GormEnginePort) *customer.Repository {
	return customer.NewRepository(repo)
}

// ProvideCustomerRepositoryPort adapta *customer.Repository a la interfaz customer.RepositoryPort.
func ProvideCustomerRepositoryPort(r *customer.Repository) customer.RepositoryPort {
	return r
}

// ProvideCustomerUseCases agrupa repositorios en customer.UseCases.
func ProvideCustomerUseCases(repo customer.RepositoryPort) *customer.UseCases {
	return customer.NewUseCases(repo)
}

// ProvideCustomerUseCasesPort adapta *customer.UseCases a la interfaz customer.UseCasesPort.
func ProvideCustomerUseCasesPort(uc *customer.UseCases) customer.UseCasesPort {
	return uc
}

// ProvideCustomerHandler construye el handler HTTP para Customer.
func ProvideCustomerHandler(
	server customer.GinEnginePort,
	useCases customer.UseCasesPort,
	cfg customer.ConfigAPIPort,
	middlewares customer.MiddlewaresEnginePort,
) *customer.Handler {
	return customer.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideCustomerConfigAPI extrae la configuración específica de API para Customer.
func ProvideCustomerConfigAPI(cfg *config.Config) customer.ConfigAPIPort {
	return &cfg.API
}

// ProvideCustomerGormEnginePort adapta *pgorm.Repository a customer.GormEnginePort.
func ProvideCustomerGormEnginePort(r *pgorm.Repository) customer.GormEnginePort {
	return r
}

// ProvideCustomerGinEnginePort adapta *pgin.Server a customer.GinEnginePort.
func ProvideCustomerGinEnginePort(s *pgin.Server) customer.GinEnginePort {
	return s
}

// ProvideCustomerMiddlewaresEnginePort adapta *mwr.Middlewares a customer.MiddlewaresEnginePort.
func ProvideCustomerMiddlewaresEnginePort(m *mwr.Middlewares) customer.MiddlewaresEnginePort {
	return m
}

// CustomerSet expone todos los providers necesarios para Customer.
var CustomerSet = wire.NewSet(
	ProvideCustomerRepository,
	ProvideCustomerRepositoryPort,
	ProvideCustomerUseCases,
	ProvideCustomerUseCasesPort,
	ProvideCustomerHandler,
	ProvideCustomerConfigAPI,
	ProvideCustomerGormEnginePort,
	ProvideCustomerGinEnginePort,
	ProvideCustomerMiddlewaresEnginePort,
)
