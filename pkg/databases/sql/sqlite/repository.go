package pkgsqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository interface {
	Connect(Config) error
	Close()
	DB() *sql.DB
	SelectContext(context.Context, any, string, ...any) error
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

var (
	instance  Repository
	once      sync.Once
	initError error
)

type repository struct {
	db     *sql.DB
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
		}
	})
	return instance, initError
}

func (r *repository) Connect(c Config) error {
	db, err := ConnectDB(c.DNS())
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	r.db = db
	return nil
}

func (r *repository) Close() {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			log.Printf("Error closing SQLite connection: %v", err)
		}
	}
}

func (r *repository) DB() *sql.DB {
	return r.db
}

// SelectContext implementa un escáner genérico para consultas SELECT
func (r *repository) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	// Obtener el tipo de destino y crear un slice para almacenar resultados
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	sliceVal := v.Elem()
	elementType := sliceVal.Type().Elem()

	// Procesar las filas
	for rows.Next() {
		// Crear un nuevo elemento del tipo correcto
		newElement := reflect.New(elementType).Elem()

		// Obtener los campos para escanear
		fields := getFields(newElement)
		scanDest := make([]any, len(fields))
		for i := range fields {
			scanDest[i] = fields[i].Addr().Interface()
		}

		// Escanear la fila en los campos
		if err := rows.Scan(scanDest...); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		// Agregar el elemento al slice
		sliceVal.Set(reflect.Append(sliceVal, newElement))
	}

	return rows.Err()
}

func (r *repository) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return r.db.QueryRowContext(ctx, query, args...)
}

func ConnectDB(connString string) (*sql.DB, error) {
	// Agregar parámetros de optimización a la cadena de conexión
	if !strings.Contains(connString, "?") {
		connString += "?"
	} else {
		connString += "&"
	}
	connString += "_journal=WAL" + // Usar Write-Ahead Logging para mejor concurrencia
		"&_sync=NORMAL" + // Balance entre rendimiento y seguridad
		"&cache=shared" + // Compartir caché entre conexiones
		"&_busy_timeout=5000" // Timeout para bloqueos (5 segundos)

	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configurar la base de datos
	db.SetMaxOpenConns(1) // SQLite es mejor con una única conexión
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Configuraciones de optimización
	pragmas := []string{
		"PRAGMA synchronous = NORMAL",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 30000000000",
		"PRAGMA page_size = 4096",
		"PRAGMA cache_size = -2000", // 2MB de caché
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}

	// Aplicar PRAGMAs
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("error setting pragma %s: %w", pragma, err)
		}
	}

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

// Función auxiliar para obtener campos de una estructura
func getFields(v reflect.Value) []reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var fields []reflect.Value
	for i := 0; i < v.NumField(); i++ {
		fields = append(fields, v.Field(i))
	}
	return fields
}
