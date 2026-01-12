package pkggorm

import "gorm.io/gorm"

// Repository es la interfaz para manejar operaciones relacionadas con GORM
type Repository interface {
	Connect(Config) error
	Client() *gorm.DB
	Address() string
	AutoMigrate(models ...any) error
}
