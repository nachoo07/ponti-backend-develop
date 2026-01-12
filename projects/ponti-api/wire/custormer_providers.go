package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
)

func ProvideCustomerRepository(repo gorm.Repository) (customer.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return customer.NewRepository(repo), nil
}

func ProvideCustomerUseCases(repo customer.Repository) customer.UseCases {
	return customer.NewUseCases(repo)
}

func ProvideCustomerHandler(server ginsrv.Server, usecases customer.UseCases, middlewares *mdw.Middlewares) *customer.Handler {
	return customer.NewHandler(server, usecases, middlewares)
}
