package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

// ProvideProjectRepository creates a Project repository instance.
func ProvideProjectRepository(repo gorm.Repository) (project.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return project.NewRepository(repo), nil
}

// ProvideProjectUseCases wires the Project use cases with its repository and required services.
func ProvideProjectUseCases(
	repo project.Repository,
	customerUC customer.UseCases,
	managerUC manager.UseCases,
	investorUC investor.UseCases,
	fieldUC field.UseCases,
	lotUC lot.UseCases,
) project.UseCases {
	return project.NewUseCases(repo, customerUC, managerUC, investorUC, fieldUC, lotUC)
}

// ProvideProjectHandler creates the HTTP handler for Project endpoints.
func ProvideProjectHandler(
	server ginsrv.Server,
	projUC project.UseCases,
	middlewares *mdw.Middlewares,
) *project.Handler {
	return project.NewHandler(server, projUC, middlewares)
}
