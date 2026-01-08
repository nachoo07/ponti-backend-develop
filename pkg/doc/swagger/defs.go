package pkgswagger

import "net/http"

// HandlerConfig define la configuración para un handler de Swagger.
type HandlerConfig struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

// Service define la interfaz para el servicio de Swagger.
type Service interface {
	// Setup configura Swagger en el enrutador proporcionado
	Setup(AddRoute func(HandlerConfig)) error
	GetConfig() Config
}

// Config define la interfaz para la configuración de Swagger.
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
