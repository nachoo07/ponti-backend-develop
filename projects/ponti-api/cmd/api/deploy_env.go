package main

import (
	"context"
	"log"

	pkgenv "github.com/alphacodinggroup/ponti-backend/pkg/config/env"
	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

func setDeployEnv(ctx context.Context, deps *wire.Dependencies) {
	platform := pkgenv.GetPlatformFromString(deps.Config.Deploy.Platform)
	env := pkgenv.GetEnvFromString(deps.Config.Deploy.Environment)

	switch platform {
	case pkgenv.Local, pkgenv.GCP:
		switch env {
		case pkgenv.Dev:
			if err := runMigrations(deps.Config.DB, deps.Config.Migrations); err != nil {
				log.Fatalf("Failed to run SQL migrations: %v", err)
			}
		}
	case pkgenv.Mix:
		switch env {
		case pkgenv.Dev:
			if err := runGormMigrations(ctx, deps.GormRepo); err != nil {
				log.Fatalf("Failed to run Gorm migrations: %v", err)
			}
		case pkgenv.Stg:
			if err := runMigrationsWithInstance(deps.GormRepo.GetSQLDB(), deps.Config.DB, deps.Config.Migrations); err != nil {
				log.Fatalf("Failed to run SQL migrations: %v", err)
			}
		default:
			log.Fatalf("Unsupported environment: %s", env)
		}

	/*case pkgenv.GCP:
	switch env {
	case pkgenv.Dev, pkgenv.Stg, pkgenv.Prod:
		if err := runMigrationsWithInstance(deps.GormRepo.GetSQLDB(), deps.Config.DB, deps.Config.Migrations); err != nil {
			log.Fatalf("Failed to run SQL migrations: %v", err)
		}
	default:
		log.Fatalf("Unsupported environment: %s", env)
	}*/

	default:
		log.Fatalf("Unsupported platform: %s", platform)
	}
}
