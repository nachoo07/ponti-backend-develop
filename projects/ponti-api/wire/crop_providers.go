package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
)

// ProvideCropRepository crea la implementación concreta de crop.Repository.
func ProvideCropRepository(repo crop.GormEnginePort) *crop.Repository {
	return crop.NewRepository(repo)
}

// ProvideCropRepositoryPort adapta *crop.Repository a la interfaz crop.RepositoryPort.
func ProvideCropRepositoryPort(r *crop.Repository) crop.RepositoryPort {
	return r
}

// ProvideCropUseCases agrupa repositorios en crop.UseCases.
func ProvideCropUseCases(
	rep crop.RepositoryPort,
) *crop.UseCases {
	return crop.NewUseCases(rep)
}

// ProvideCropUseCasesPort adapta *crop.UseCases a la interfaz crop.UseCasesPort.
func ProvideCropUseCasesPort(uc *crop.UseCases) crop.UseCasesPort {
	return uc
}

// ProvideCropHandler construye el handler HTTP para Crop.
func ProvideCropHandler(
	server crop.GinEnginePort,
	useCases crop.UseCasesPort,
	cfg crop.ConfigAPIPort,
	middlewares crop.MiddlewaresEnginePort,
) *crop.Handler {
	return crop.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideCropConfigAPI extrae la configuración específica de API para Crop.
func ProvideCropConfigAPI(cfg *config.Config) crop.ConfigAPIPort {
	return &cfg.API
}

// ProvideCropGormEnginePort adapta *pgorm.Repository a crop.GormEnginePort.
func ProvideCropGormEnginePort(r *pgorm.Repository) crop.GormEnginePort {
	return r
}

// ProvideCropGinEnginePort adapta *pgin.Server a crop.GinEnginePort.
func ProvideCropGinEnginePort(s *pgin.Server) crop.GinEnginePort {
	return s
}

// ProvideCropMiddlewaresEnginePort adapta *mwr.Middlewares a crop.MiddlewaresEnginePort.
func ProvideCropMiddlewaresEnginePort(m *mwr.Middlewares) crop.MiddlewaresEnginePort {
	return m
}

// CropSet expone todos los providers necesarios para Crop.
var CropSet = wire.NewSet(
	ProvideCropRepository,
	ProvideCropRepositoryPort,
	ProvideCropUseCases,
	ProvideCropUseCasesPort,
	ProvideCropHandler,
	ProvideCropConfigAPI,
	ProvideCropGormEnginePort,
	ProvideCropGinEnginePort,
	ProvideCropMiddlewaresEnginePort,
)
