package pkgpg

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq" // Importación de driver
)

type Repository interface {
	Connect(Config) error
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

func newRepository(c Config) (Repository, error) {
	once.Do(func() {
		instance = &repository{}
		initError = instance.Connect(c)
		if initError != nil {
			instance = nil
		}
	})
	return instance, initError
}

func (r *repository) Connect(c Config) error {
	// Construir la cadena de conexión
	connString := c.DNS()

	// Conectar con la base de datos PostgreSQL
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Verificar la conexión
	if err = db.Ping(); err != nil {
		return fmt.Errorf("unable to ping the database: %w", err)
	}

	r.db = db
	return nil
}

func (r *repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *repository) DB() *sql.DB {
	return r.db
}
