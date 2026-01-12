package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
)

func ProvideManagerRepository(repo gorm.Repository) (manager.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return manager.NewRepository(repo), nil
}

func ProvideManagerUseCases(repo manager.Repository) manager.UseCases {
	return manager.NewUseCases(repo)
}

func ProvideManagerHandler(server ginsrv.Server, usecases manager.UseCases, middlewares *mdw.Middlewares) *manager.Handler {
	return manager.NewHandler(server, usecases, middlewares)
}
