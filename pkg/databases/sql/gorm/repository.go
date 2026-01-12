package pkggorm

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// repository es la implementación de Repository
type repository struct {
	client  *gorm.DB
	address string
	config  Config
}

// NewRepository inicializa un nuevo repositorio sin usar singleton
func newRepository(c Config) (Repository, error) {
	repo := &repository{
		config: c,
	}
	if err := repo.Connect(c); err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}
	return repo, nil
}

// Connect establece la conexión con la base de datos según el tipo
func (r *repository) Connect(config Config) error {
	// Crear la base de datos si no existe
	if err := r.createDatabaseIfNotExists(config); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	var db *gorm.DB
	var err error

	switch config.GetDBType() {
	case Postgres:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			config.GetHost(), config.GetUser(), config.GetPassword(), config.GetDBName(), config.GetPort())
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}

	case MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetUser(), config.GetPassword(), config.GetHost(), config.GetPort(), config.GetDBName())
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL: %w", err)
		}

	case SQLite:
		dsn := config.GetSQLitePath()
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database type: %s", config.GetDBType())
	}

	r.client = db
	r.address = config.GetHost()

	// Verifica la conexión realizando un ping (solo para MySQL y PostgreSQL)
	if config.GetDBType() != SQLite {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
	}

	log.Printf("Gorm successfully connected to %s database: %s", config.GetDBType(), config.GetDBName())
	return nil
}

// Métodos de `repository` para implementar la interfaz Repository

func (r *repository) Client() *gorm.DB {
	return r.client
}

func (r *repository) Address() string {
	return r.address
}

func (r *repository) AutoMigrate(models ...any) error {
	return r.client.AutoMigrate(models...)
}

func (r *repository) createDatabaseIfNotExists(config Config) error {
	switch config.GetDBType() {
	case Postgres:
		// Conecta al servidor usando la base de datos predeterminada "postgres"
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable",
			config.GetHost(), config.GetUser(), config.GetPassword(), config.GetPort())
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		defer sqlDB.Close()

		// Crear la base de datos si no existe
		createDBSQL := fmt.Sprintf("CREATE DATABASE %s", config.GetDBName())
		if err := db.Exec(createDBSQL).Error; err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create database: %w", err)
			}
		}

	case MySQL:
		// Conecta al servidor sin especificar la base de datos
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetUser(), config.GetPassword(), config.GetHost(), config.GetPort())
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL server: %w", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		defer sqlDB.Close()

		// Crear la base de datos si no existe
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.GetDBName())
		if err := db.Exec(createDBSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

	case SQLite:
		// Verificar si el archivo de la base de datos existe
		if _, err := os.Stat(config.GetSQLitePath()); os.IsNotExist(err) {
			fmt.Println("Automatically created by SQLite")
		}
		// SQLite creará automáticamente la base de datos cuando se conecte
		return nil

	default:
		return fmt.Errorf("unsupported database type: %s", config.GetDBType())
	}

	return nil
}
