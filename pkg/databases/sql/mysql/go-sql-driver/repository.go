package pkgmysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type Repository interface {
	Close()
	DB() *sql.DB
}

var (
	instance  Repository
	once      sync.Once
	initError error
)

type repository struct {
	db *sql.DB
}

// newRepository crea una nueva instancia de repository con configuración proporcionada.
func newRepository(c config) (Repository, error) {
	once.Do(func() {
		client := &repository{}
		initError = client.connect(c)
		if initError != nil {
			instance = nil
		} else {
			instance = client
		}
	})
	return instance, initError
}

// connect establece la conexión a la base de datos MySQL.
func (r *repository) connect(c config) error {
	dsn := c.dsn()
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}
	r.db = conn
	return nil
}

// Ping verifica la conexión a la base de datos.
func (r *repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Close cierra la conexión a la base de datos.
func (r *repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

// DB devuelve la instancia *sql.DB.
func (r *repository) DB() *sql.DB {
	return r.db
}
