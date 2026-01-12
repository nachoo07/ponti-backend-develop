package pkgpostgresql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository interface {
	Connect(Config) error
	Close()
	Pool() *pgxpool.Pool
	SelectContext(context.Context, any, string, ...any) error
	QueryRowContext(context.Context, string, ...any) pgx.Row
}

type Config interface {
	Validate() error
	DNS() string
	GetHost() string
	GetUser() string
	GetPassword() string
	GetDbName() string
	GetPort() string
	GetMigrationsDir() string
}
