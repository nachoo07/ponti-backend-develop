package pkgpostgresql

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	instance  Repository
	once      sync.Once
	initError error
)

type repository struct {
	pool   *pgxpool.Pool
	config Config
}

func newRepository(c Config) (Repository, error) {
	once.Do(func() {
		instance = &repository{
			config: c,
		}
		initError = instance.Connect(c)
		if initError != nil {
			instance = nil
		} else {
			log.Printf("Postgres successfully connected to database: %s", c.GetDbName())
		}
	})
	return instance, initError
}

func GetInstance() (Repository, error) {
	if instance == nil {
		return nil, fmt.Errorf("repository instance is not initialized")
	}
	return instance, nil
}

func (r *repository) Connect(c Config) error {
	// Construcción de la cadena de conexión
	connString := c.DNS()

	// Conexión al pool de PostgreSQL
	pool, err := ConnectPool(connString)
	if err != nil {
		return err
	}
	r.pool = pool
	return nil
}

func (r *repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *repository) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return pgxscan.Select(ctx, r.pool, dest, query, args...)
}

func ConnectPool(connString string) (*pgxpool.Pool, error) {
	// Usar context.Background() para permitir todos los reintentos
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database connection string: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 1
	config.HealthCheckPeriod = 1 * time.Minute
	config.MaxConnLifetime = 24 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	maxRetries := 5
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Postgres connection attempt %d of %d", i+1, maxRetries)

		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pingCtx)
			pingCancel()

			if err == nil {
				log.Printf("Postgres successfully connected on attempt %d", i+1)
				return pool, nil
			}
			pool.Close()
		}

		log.Printf("Attempt %d failed: %v", i+1, err)

		if i < maxRetries-1 {
			log.Printf("Waiting %v before next attempt...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

func (r *repository) QueryRowContext(ctx context.Context, query string, args ...any) pgx.Row {
	return r.pool.QueryRow(ctx, query, args...)
}
