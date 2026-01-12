package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
)

func ProvideLotRepository(repo gorm.Repository) (lot.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return lot.NewRepository(repo), nil
}

func ProvideLotUseCases(repo lot.Repository, cropUC crop.UseCases) lot.UseCases {
	return lot.NewUseCases(repo, cropUC)
}

func ProvideLotHandler(server ginsrv.Server, usecases lot.UseCases, middlewares *mdw.Middlewares) *lot.Handler {
	return lot.NewHandler(server, usecases, middlewares)
}
