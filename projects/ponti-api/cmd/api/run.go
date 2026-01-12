package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	personmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/person/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	usermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/repository/models"

	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

// RunHttpServer registers routes in the Gin router and starts the HTTP server.
func RunHttpServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	log.Println("Registering HTTP routes...")

	// Configure global middlewares if any.
	if len(deps.Middlewares.Global) > 0 {
		deps.GinServer.GetRouter().Use(deps.Middlewares.Global...)
	}

	// Register all application routes.
	log.Println("Starting HTTP Server...")
	registerHttpRoutes(deps)

	// Start the HTTP server (e.g., on port 8080).
	return deps.GinServer.RunServer(ctx)
}

// registerHttpRoutes registers all application routes in the Gin router.
func registerHttpRoutes(deps *wire.Dependencies) {
	deps.PersonHandler.Routes()
	deps.UserHandler.Routes()
	deps.NotificationHandler.Routes()
	deps.LotHandler.Routes()
	deps.CustomerHandler.Routes()
	deps.InvestorHandler.Routes()
	deps.FieldHandler.Routes()
	deps.ProjectHandler.Routes()
	deps.CropHandler.Routes()
	deps.ManagerHandler.Routes()
}

// RunGormMigrations runs SQL migrations using GORM.
func RunGormMigrations(ctx context.Context, repo gorm.Repository) error {
	log.Println("Starting GORM migrations...")

	// Obtain the underlying database connection.
	sqlDB, err := repo.Client().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// List of models to migrate (entidades existentes + 6 nuevas).
	modelsToMigrate := []any{
		&personmodels.Person{},
		&usermodels.User{},
		&usermodels.Follow{},
		&lotmodels.Lot{},
		&customermodels.Customer{},
		&investormodels.Investor{},
		&fieldmodels.Field{},
		&projectmodels.Project{},
		&cropmodels.Crop{},
		&managermodels.Manager{},
	}

	start := time.Now()
	if err := repo.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("failed to migrate database models: %w", err)
	}
	duration := time.Since(start)
	log.Printf("GORM migrations completed successfully in %s.", duration)

	return nil
}
