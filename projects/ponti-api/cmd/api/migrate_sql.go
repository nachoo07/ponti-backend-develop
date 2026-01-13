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

/*func buildMigrateDatabaseURL(cfg config.DB) string {
	user := strings.TrimSpace(cfg.User)
	pass := strings.TrimSpace(cfg.Password)
	host := strings.TrimSpace(cfg.Host)
	name := strings.TrimSpace(cfg.Name)
	ssl := strings.TrimSpace(cfg.SSLMode)

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%d", host, cfg.Port),
		Path:   "/" + name,
	}

	if ssl != "" {
		q := url.Values{}
		q.Set("sslmode", ssl)
		u.RawQuery = q.Encode()
	}

	return u.String()
}*/


func buildMigrateDatabaseURL(cfg config.DB) string {
    user := strings.TrimSpace(cfg.User)
    pass := strings.TrimSpace(cfg.Password)
    host := strings.TrimSpace(cfg.Host)
    name := strings.TrimSpace(cfg.Name)
    ssl := strings.TrimSpace(cfg.SSLMode)

    // ✅ Detectar si es un socket de Cloud SQL
    if strings.HasPrefix(host, "/") {
        // Socket de Cloud SQL: 
        // CAMBIO CLAVE: Usamos "" (vacío) en vez de "localhost"
        // para forzar a lib/pq a usar el parámetro host del query.
        u := &url.URL{
            Scheme: "postgres",
            User:   url.UserPassword(user, pass),
            Host:   "", // <--- AQUÍ ESTÁ EL TRUCO (Antes decía "localhost")
            Path:   name, // Sin barra inicial si Host está vacío, o url.URL lo maneja
        }

        // url.URL a veces necesita ayuda con el path si el host es vacío
        // Para asegurar que quede "postgres://user:pass@/dbname", forzamos el path:
        u.Path = "/" + name

        q := url.Values{}
        q.Set("host", host) // Esto inyecta /cloudsql/...
        if ssl != "" {
            q.Set("sslmode", ssl)
        }
        u.RawQuery = q.Encode()

        return u.String()
    }

    // TCP tradicional (local/remoto)
    u := &url.URL{
        Scheme: "postgres",
        User:   url.UserPassword(user, pass),
        Host:   fmt.Sprintf("%s:%d", host, cfg.Port),
        Path:   "/" + name,
    }

    if ssl != "" {
        q := url.Values{}
        q.Set("sslmode", ssl)
        u.RawQuery = q.Encode()
    }

    return u.String()
}