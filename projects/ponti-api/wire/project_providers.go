package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"

	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

// ProvideProjectGormEnginePort ...
func ProvideProjectGormEnginePort(r *gormpkg.Repository) project.GormEnginePort {
	return r
}

// ProvideProjectRepository ...
func ProvideProjectRepository(r project.GormEnginePort) *project.Repository {
	return project.NewRepository(r)
}

// ProvideProjectRepositoryPort ...
func ProvideProjectRepositoryPort(repo *project.Repository) project.RepositoryPort {
	return repo
}

// ProvideProjectConfigAPI ...
func ProvideProjectConfigAPI(c *cfg.Config) project.ConfigAPIPort {
	return &c.API
}

// ProvideProjectGinEnginePort ...
func ProvideProjectGinEnginePort(s *pgin.Server) project.GinEnginePort {
	return s
}

// ProvideProjectMiddlewaresEnginePort ...
func ProvideProjectMiddlewaresEnginePort(m *mwr.Middlewares) project.MiddlewaresEnginePort {
	return m
}

// ProvideProjectSuggesterPort ...
func ProvideProjectSuggesterPort(s *sug.WordsSuggester) project.WordsSuggesterPort {
	return project.NewWordsSuggester(s)
}

// ProvideProjectUseCases ...
func ProvideProjectUseCases(rep project.RepositoryPort, sug project.WordsSuggesterPort) *project.UseCases {
	return project.NewUseCases(rep, sug)
}

// ProvideProjectUseCasesPort ...
func ProvideProjectUseCasesPort(u *project.UseCases) project.UseCasesPort {
	return u
}

// ProvideProjectHandler ...
func ProvideProjectHandler(
	server project.GinEnginePort,
	ucs project.UseCasesPort,
	cfg project.ConfigAPIPort,
	mws project.MiddlewaresEnginePort,
) *project.Handler {
	return project.NewHandler(ucs, server, cfg, mws)
}

// ProjectSet ...
var ProjectSet = wire.NewSet(
	ProvideProjectGormEnginePort,
	ProvideProjectRepository,
	ProvideProjectRepositoryPort,
	ProvideProjectConfigAPI,
	ProvideProjectGinEnginePort,
	ProvideProjectMiddlewaresEnginePort,

	ProvideProjectSuggesterPort,

	ProvideProjectUseCases,
	ProvideProjectUseCasesPort,
	ProvideProjectHandler,
)
