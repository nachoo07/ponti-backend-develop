package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	classtype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype" // Corrected import for classtype
)

// ProvideClassTypeGormEnginePort ...
func ProvideClassTypeGormEnginePort(r *gormpkg.Repository) classtype.GormEnginePort {
	return r
}
func ProvideClassTypeRepository(r classtype.GormEnginePort) *classtype.Repository {
	return classtype.NewRepository(r)
}
func ProvideClassTypeRepositoryPort(repo *classtype.Repository) classtype.RepositoryPort {
	return repo
}

// ProvideClassTypeConfigAPI ...
func ProvideClassTypeConfigAPI(c *cfg.Config) classtype.ConfigAPIPort {
	return &c.API
}

// ProvideClassTypeGinEnginePort ...
func ProvideClassTypeGinEnginePort(s *pgin.Server) classtype.GinEnginePort {
	return s
}
func ProvideClassTypeMiddlewaresEnginePort(m *mwr.Middlewares) classtype.MiddlewaresEnginePort {
	return m
}

// ProvideClassTypeUseCases ...
func ProvideClassTypeUseCases(rep classtype.RepositoryPort) *classtype.UseCases {
	return classtype.NewUseCases(rep)
}

// ProvideClassTypeUseCasesPort ...
func ProvideClassTypeUseCasesPort(u *classtype.UseCases) classtype.UseCasesPort {
	return u
}

// ProvideClassTypeHandler ...
func ProvideClassTypeHandler(
	server classtype.GinEnginePort,
	ucs classtype.UseCasesPort,
	cfg classtype.ConfigAPIPort,
	mws classtype.MiddlewaresEnginePort,
) *classtype.Handler {
	return classtype.NewHandler(ucs, server, cfg, mws)
}

// ClassTypeSet ...
var ClassTypeSet = wire.NewSet(
	ProvideClassTypeGormEnginePort,
	ProvideClassTypeRepository,
	ProvideClassTypeRepositoryPort,
	ProvideClassTypeConfigAPI,
	ProvideClassTypeGinEnginePort,
	ProvideClassTypeMiddlewaresEnginePort,

	ProvideClassTypeUseCases,
	ProvideClassTypeUseCasesPort,
	ProvideClassTypeHandler,
)
