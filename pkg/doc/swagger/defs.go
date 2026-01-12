package pkgswagger

import "net/http"

// HandlerConfig contiene la configuración para un manejador HTTP
type HandlerConfig struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

// Service define las operaciones disponibles para el servicio Swagger
type Service interface {
	// Setup configura Swagger en el router proporcionado
	Setup(AddRoute func(HandlerConfig)) error
	GetConfig() Config
}

// Config define la configuración necesaria para Swagger
type Config interface {
	GetTitle() string
	GetDescription() string
	GetVersion() string
	GetHost() string
	GetBasePath() string
	GetSchemes() []string
	IsEnabled() bool
	Validate() error
}
