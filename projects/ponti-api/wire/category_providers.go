package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	category "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category" // Corrected import for category
)

// ProvideCategoryGormEnginePort ...
func ProvideCategoryGormEnginePort(r *gormpkg.Repository) category.GormEnginePort {
	return r
}
func ProvideCategoryRepository(r category.GormEnginePort) *category.Repository {
	return category.NewRepository(r)
}
func ProvideCategoryRepositoryPort(repo *category.Repository) category.RepositoryPort {
	return repo
}

// ProvideCategoryConfigAPI ...
func ProvideCategoryConfigAPI(c *cfg.Config) category.ConfigAPIPort {
	return &c.API
}

// ProvideCategoryGinEnginePort ...
func ProvideCategoryGinEnginePort(s *pgin.Server) category.GinEnginePort {
	return s
}
func ProvideCategoryMiddlewaresEnginePort(m *mwr.Middlewares) category.MiddlewaresEnginePort {
	return m
}

// ProvideCategoryUseCases ...
func ProvideCategoryUseCases(rep category.RepositoryPort) *category.UseCases {
	return category.NewUseCases(rep)
}

func ProvideCategoryUseCasesPort(u *category.UseCases) category.UseCasesPort {
	return u
}

// ProvideCategoryHandler ...
func ProvideCategoryHandler(
	server category.GinEnginePort,
	ucs category.UseCasesPort,
	cfg category.ConfigAPIPort,
	mws category.MiddlewaresEnginePort,
) *category.Handler {
	return category.NewHandler(ucs, server, cfg, mws)
}

// CategorySet ...
var CategorySet = wire.NewSet(
	ProvideCategoryGormEnginePort,
	ProvideCategoryRepository,
	ProvideCategoryRepositoryPort,
	ProvideCategoryConfigAPI,
	ProvideCategoryGinEnginePort,
	ProvideCategoryMiddlewaresEnginePort,

	ProvideCategoryUseCases,
	ProvideCategoryUseCasesPort,
	ProvideCategoryHandler,
)
