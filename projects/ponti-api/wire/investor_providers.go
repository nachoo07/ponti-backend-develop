package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
)

func ProvideInvestorRepository(repo gorm.Repository) (investor.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return investor.NewRepository(repo), nil
}

func ProvideInvestorUseCases(repo investor.Repository) investor.UseCases {
	return investor.NewUseCases(repo)
}

func ProvideInvestorHandler(server ginsrv.Server, usecases investor.UseCases, middlewares *mdw.Middlewares) *investor.Handler {
	return investor.NewHandler(server, usecases, middlewares)
}
