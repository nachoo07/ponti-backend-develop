package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
	m, err := migrate.New(
		migConfig.Dir,
		buildMigrateDatabaseURL(dbConfig),
	)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error applying migrations: %w", err)
	}
	return nil
}

func runMigrationsWithInstance(sqlDB *sql.DB, dbConfig config.DB, migConfig config.Migrations) error {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migConfig.Dir,
		dbConfig.Name,
		driver,
	)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

func buildMigrateDatabaseURL(cfg config.DB) string {
    user := strings.TrimSpace(cfg.User)
    pass := strings.TrimSpace(cfg.Password)
    host := strings.TrimSpace(cfg.Host)
    name := strings.TrimSpace(cfg.Name)
    ssl := strings.TrimSpace(cfg.SSLMode)

    // Preparamos los Query Params
    q := url.Values{}
    if ssl != "" {
        q.Set("sslmode", ssl)
    }

    var urlHost string

    // 🚨 CORRECCIÓN AQUÍ 🚨
    // Detectamos si es un Socket Unix (empieza con /) típico de Cloud SQL
    if strings.HasPrefix(host, "/") {
        // Si es socket, el 'Host' de la URL debe ser genérico (ej: localhost)
        // y la ruta real del socket se pasa como parámetro ?host=...
        urlHost = "localhost" 
        q.Set("host", host)
    } else {
        // Si es conexión normal (TCP/IP), usamos host:port
        urlHost = fmt.Sprintf("%s:%d", host, cfg.Port)
    }

    u := &url.URL{
        Scheme:   "postgres",
        User:     url.UserPassword(user, pass),
        Host:     urlHost,
        Path:     "/" + name,
        RawQuery: q.Encode(), // Esto añade ?sslmode=...&host=/cloudsql/...
    }

    return u.String()
}
