package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	campaign "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign"
)

// ProvideCampaignRepository crea la implementación concreta de campaign.Repository.
func ProvideCampaignRepository(repo campaign.GormEnginePort) *campaign.Repository {
	return campaign.NewRepository(repo)
}

// ProvideCampaignRepositoryPort adapta *campaign.Repository a la interfaz campaign.RepositoryPort.
func ProvideCampaignRepositoryPort(r *campaign.Repository) campaign.RepositoryPort {
	return r
}

// ProvideCampaignUseCases agrupa repositorio en campaign.UseCases.
func ProvideCampaignUseCases(
	rep campaign.RepositoryPort,
) *campaign.UseCases {
	return campaign.NewUseCases(rep)
}

// ProvideCampaignUseCasesPort adapta *campaign.UseCases a la interfaz campaign.UseCasesPort.
func ProvideCampaignUseCasesPort(uc *campaign.UseCases) campaign.UseCasesPort {
	return uc
}

// ProvideCampaignHandler construye el handler HTTP para Campaign.
func ProvideCampaignHandler(
	server campaign.GinEnginePort,
	useCases campaign.UseCasesPort,
	cfg campaign.ConfigAPIPort,
	middlewares campaign.MiddlewaresEnginePort,
) *campaign.Handler {
	return campaign.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideCampaignConfigAPI extrae la configuración específica de API para Campaign.
func ProvideCampaignConfigAPI(cfg *config.Config) campaign.ConfigAPIPort {
	return &cfg.API
}

// ProvideCampaignGormEnginePort adapta *pgorm.Repository a campaign.GormEnginePort.
func ProvideCampaignGormEnginePort(r *pgorm.Repository) campaign.GormEnginePort {
	return r
}

// ProvideCampaignGinEnginePort adapta *pgin.Server a campaign.GinEnginePort.
func ProvideCampaignGinEnginePort(s *pgin.Server) campaign.GinEnginePort {
	return s
}

// ProvideCampaignMiddlewaresEnginePort adapta *mwr.Middlewares a campaign.MiddlewaresEnginePort.
func ProvideCampaignMiddlewaresEnginePort(m *mwr.Middlewares) campaign.MiddlewaresEnginePort {
	return m
}

// CampaignSet expone todos los providers necesarios para Campaign.
var CampaignSet = wire.NewSet(
	ProvideCampaignRepository,
	ProvideCampaignRepositoryPort,
	ProvideCampaignUseCases,
	ProvideCampaignUseCasesPort,
	ProvideCampaignHandler,
	ProvideCampaignConfigAPI,
	ProvideCampaignGormEnginePort,
	ProvideCampaignGinEnginePort,
	ProvideCampaignMiddlewaresEnginePort,
)
