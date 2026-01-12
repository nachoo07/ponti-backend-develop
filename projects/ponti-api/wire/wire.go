//go:build wireinject
// +build wireinject

package wire

import (
	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/postgresql/pgxpool"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	smtp "github.com/alphacodinggroup/ponti-backend/pkg/notification/smtp"
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	notification "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/notification"
	person "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	user "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user"
)

type Dependencies struct {
	ConfigLoader       config.Loader
	GinServer          ginsrv.Server
	GormRepository     gorm.Repository
	PostgresRepository pg.Repository
	SmtpService        smtp.Service

	Middlewares *mdw.Middlewares

	PersonHandler       *person.Handler
	UserHandler         *user.Handler
	NotificationHandler *notification.Handler
	CropHandler         *crop.Handler
	CustomerHandler     *customer.Handler
	ManagerHandler      *manager.Handler
	FieldHandler        *field.Handler
	InvestorHandler     *investor.Handler
	LotHandler          *lot.Handler
	ProjectHandler      *project.Handler

	PersonUseCases   person.UseCases
	UserUseCases     user.UseCases
	CropUseCases     crop.UseCases
	CustomerUseCases customer.UseCases
	FieldUseCases    field.UseCases
	InvestorUseCases investor.UseCases
	LotUseCases      lot.UseCases
	ProjectUseCases  project.UseCases
}

func Initialize() (*Dependencies, error) {
	wire.Build(
		ProvideConfigLoader,
		ProvideGinServer,
		ProvideGormRepository,
		ProvidePostgresRepository,
		ProvideJwtMiddleware,
		ProvideMiddlewares,
		ProvideSmtpService,

		ProvidePersonRepository,
		ProvidePersonUseCases,
		ProvidePersonHandler,

		ProvideUserRepository,
		ProvideUserUseCases,
		ProvideUserHandler,

		ProvideNotificationSmtpService,
		ProvideNotificationUseCases,
		ProvideNotificationHandler,

		ProvideCropRepository,
		ProvideCropUseCases,
		ProvideCropHandler,

		ProvideCustomerRepository,
		ProvideCustomerUseCases,
		ProvideCustomerHandler,

		ProvideManagerRepository,
		ProvideManagerUseCases,
		ProvideManagerHandler,

		ProvideFieldRepository,
		ProvideFieldUseCases,
		ProvideFieldHandler,

		ProvideInvestorRepository,
		ProvideInvestorUseCases,
		ProvideInvestorHandler,

		ProvideLotRepository,
		ProvideLotUseCases,
		ProvideLotHandler,

		ProvideProjectRepository,
		ProvideProjectUseCases,
		ProvideProjectHandler,

		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
