package wire

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

// --- GORM ----

type GormEnginePort interface {
	Address() string
	AutoMigrate(...any) error
	Client() *gorm.DB
	Connect(pgorm.ConfigPort) error
}

func ProvideGormRepository(cfg *config.DB) (*pgorm.Repository, error) {
	return pgorm.Bootstrap(
		cfg.Type,
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
		cfg.Port,
	)
}

func ProvideGormRepositoryPort(repo *pgorm.Repository) GormEnginePort {
	return repo
}

var GormSet = wire.NewSet(
	ProvideGormRepository,
	ProvideGormRepositoryPort,
)

// --- Gin Server ----

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
	WrapH(h http.Handler) gin.HandlerFunc
}

func ProvideGinEngine(cfg *config.Config) (*pgin.Server, error) {
	return pgin.Bootstrap(
		strconv.Itoa(cfg.HTTPServer.Port),
		cfg.API.Version,
		false,
	)
}

func ProvideGinEnginePort(srv *pgin.Server) GinEnginePort {
	return srv
}

var GinSet = wire.NewSet(
	ProvideGinEngine,
	ProvideGinEnginePort,
)

// --- WordsSuggester Providers -------------------------------------------

type WordsSuggesterEnginePort interface {
	Suggest(context.Context, string, string, string, int, int) ([]sug.Suggestion, int64, error)
	Close() error
	Health(ctx context.Context) error
}

func ProvideSuggesterDB(repo *pgorm.Repository) sug.DB {
	return sug.NewPkggormAdapter(repo)
}

func ProvideSuggester(db sug.DB, cfg *config.WordsSuggester) (*sug.WordsSuggester, error) {
	return sug.Bootstrap(
		sug.WithDB(db),
		sug.WithLimit(cfg.Limit),
		sug.WithThreshold(cfg.Threshold),
	)
}

func ProvideSuggesterEnginePort(s *sug.WordsSuggester) WordsSuggesterEnginePort {
	return s
}

var SuggesterSet = wire.NewSet(
	ProvideSuggesterDB,
	ProvideSuggester,
	ProvideSuggesterEnginePort,
)
