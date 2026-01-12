package main

import (
	"context"
	"errors"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

// runHttpServer registers routes in the Gin router and starts the HTTP server.
func runHttpServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	// Set up the Gin router with global middlewares only
	// Global middlewares: ErrorHandling, RequestAndResponseLogger
	deps.GinEngine.GetRouter().Use(deps.Middlewares.GetGlobal()...)

	// Register all application routes.
	// Each handler will apply its own validation middlewares as needed
	registerHttpRoutes(deps)

	log.Println("Starting HTTP Server on port: ", deps.Config.HTTPServer.Port)
	log.Println("Version: ", deps.Config.App.Version)
	log.Println("--------------------------------")
	log.Println("Database: ", deps.Config.DB.Name)
	log.Println("--------------------------------")
	log.Println("Environment: ", deps.Config.Deploy.Environment)
	log.Println("Platform: ", deps.Config.Deploy.Platform)

	// Start the HTTP server (e.g., on port 8080).
	return deps.GinEngine.RunServer(ctx)
}

// registerHttpRoutes registers all application routes in the Gin router.
func registerHttpRoutes(deps *wire.Dependencies) {
	deps.LotHandler.Routes()
	deps.CustomerHandler.Routes()
	deps.CampaignHandler.Routes()
	deps.DashboardHandler.Routes()
	deps.DataIntegrityHandler.Routes()
	deps.InvestorHandler.Routes()
	deps.InvoiceHandler.Routes()
	deps.FieldHandler.Routes()
	deps.ProjectHandler.Routes()
	deps.ReportHandler.Routes()
	deps.CropHandler.Routes()
	deps.ManagerHandler.Routes()
	deps.LeaseTypeHandler.Routes()
	deps.SupplyHandler.Routes()
	deps.CategoryHandler.Routes()
	deps.ClassTypeHandler.Routes()
	deps.AppParametersHandler.Routes()
	deps.WorkorderHandler.Routes()
	deps.DollarHandler.Routes()
	deps.LaborHandler.Routes()
	deps.StockHandler.Routes()
	deps.CommercializationHandler.Routes()
	deps.SupplyMovement.Routes()
}
