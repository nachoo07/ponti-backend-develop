package wire

import (
	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/google/wire"

	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	invoice "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice"
)

// ---- GORM -----

func ProvideInvoiceGormEnginePort(r *gormpkg.Repository) invoice.GormEnginePort {
	return r
}

func ProvideInvoiceRepository(r invoice.GormEnginePort) *invoice.Repository {
	return invoice.NewRepository(r)
}

func ProvideInvoiceRepositoryPort(r *invoice.Repository) invoice.RepositoryPort {
	return r
}

// ---- CONFIG API ----

func ProvideInvoiceConfigAPI(c *cfg.Config) invoice.ConfigAPIPort {
	return &c.API
}

// ---- HTTP & MIDDLEWARE ---

func ProvideInvoiceGinEnginePort(s *pgin.Server) invoice.GinEnginePort {
	return s
}

func ProvideInvoiceMiddlewaresEnginePort(m *mwr.Middlewares) invoice.MiddlewaresEnginePort {
	return m
}

// ---- USE CASES ----

func ProvideInvoiceUseCases(r invoice.RepositoryPort) *invoice.UseCases {
	return invoice.NewUseCases(r)
}

func ProvideInvoiceUseCasePort(u *invoice.UseCases) invoice.UseCasePort {
	return u
}

// ---- HANDLER ----

func ProvideInvoiceHandler(
	server invoice.GinEnginePort,
	ucs invoice.UseCasePort,
	cfg invoice.ConfigAPIPort,
	mws invoice.MiddlewaresEnginePort,
) *invoice.Handler {
	return invoice.NewHandler(ucs, server, cfg, mws)
}

var InvoiceSet = wire.NewSet(
	ProvideInvoiceGormEnginePort,
	ProvideInvoiceRepository,
	ProvideInvoiceRepositoryPort,

	ProvideInvoiceConfigAPI,

	ProvideInvoiceGinEnginePort,
	ProvideInvoiceMiddlewaresEnginePort,

	ProvideInvoiceUseCases,
	ProvideInvoiceUseCasePort,

	ProvideInvoiceHandler,
)
