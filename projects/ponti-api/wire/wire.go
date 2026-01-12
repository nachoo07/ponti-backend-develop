//go:build wireinject
// +build wireinject

package wire

import (
	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	app_parameters "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/app_parameters"
	campaign "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign"
	category "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category"
	classtype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype"
	commercialization "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization"
	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	dashboard "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard"
	data_integrity "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/data-integrity"
	dollar "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	invoice "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice"
	labor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	report "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	supply "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement"
	workorder "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder"
	"github.com/google/wire"
)

type Dependencies struct {
	Config                   *config.Config
	GinEngine                *gin.Server
	GormRepo                 *gorm.Repository
	Middlewares              *mwr.Middlewares
	WordsSuggester           *sug.WordsSuggester
	CustomerHandler          *customer.Handler
	CampaignHandler          *campaign.Handler
	DashboardHandler         *dashboard.Handler
	DataIntegrityHandler     *data_integrity.Handler
	InvestorHandler          *investor.Handler
	CropHandler              *crop.Handler
	LotHandler               *lot.Handler
	FieldHandler             *field.Handler
	ManagerHandler           *manager.Handler
	ProjectHandler           *project.Handler
	ReportHandler            *report.ReportHandler
	LeaseTypeHandler         *leasetype.Handler
	SupplyHandler            *supply.Handler
	CategoryHandler          *category.Handler
	AppParametersHandler     *app_parameters.Handler
	ClassTypeHandler         *classtype.Handler
	DollarHandler            *dollar.Handler
	WorkorderHandler         *workorder.Handler
	LaborHandler             *labor.Handler
	InvoiceHandler           *invoice.Handler
	CommercializationHandler *commercialization.Handler
	StockHandler             *stock.Handler
	SupplyMovement           *supply_movement.Handler
}

func Initialize() (*Dependencies, error) {
	wire.Build(
		ConfigSet,
		GormSet,
		GinSet,
		MiddlewareSet,
		SuggesterSet,
		CustomerSet,
		CampaignSet,
		DashboardSet,
		DataIntegritySet,
		InvestorSet,
		CropSet,
		CommercializationSet,
		LotSet,
		FieldSet,
		ManagerSet,
		ProjectSet,
		ReportSet,
		LeaseTypeSet,
		SupplySet,
		CategorySet,
		AppParametersSet,
		ClassTypeSet,
		DollarSet,
		WorkorderSet,
		LaborSet,
		StockSet,
		InvoiceSet,
		SupplyMovementSet,
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
