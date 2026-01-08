package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	report "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report"
)

// ProvideReportRepository crea la implementación concreta de report.Repository.
func ProvideReportRepository(repo report.GormEnginePort) *report.ReportRepository {
	return report.NewReportRepository(repo)
}

// ProvideReportRepositoryPort adapta *report.ReportRepository a la interfaz report.ReportRepositoryPort.
func ProvideReportRepositoryPort(r *report.ReportRepository) report.ReportRepositoryPort {
	return r
}

// ProvideReportUseCases agrupa repositorio en report.UseCases.
func ProvideReportUseCases(
	rep report.ReportRepositoryPort,
) *report.ReportUseCase {
	return report.NewReportUseCase(rep)
}

// ProvideReportUseCasesPort adapta *report.ReportUseCase a la interfaz report.ReportUseCasePort.
func ProvideReportUseCasesPort(uc *report.ReportUseCase) report.ReportUseCasePort {
	return uc
}

// ProvideReportHandler construye el handler HTTP para Report.
func ProvideReportHandler(
	server report.GinEnginePort,
	useCases report.ReportUseCasePort,
	cfg report.ConfigAPIPort,
	middlewares report.MiddlewaresEnginePort,
) *report.ReportHandler {
	return report.NewReportHandler(useCases, server, cfg, middlewares)
}

// ProvideReportConfigAPI extrae la configuración específica de API para Report.
func ProvideReportConfigAPI(cfg *config.Config) report.ConfigAPIPort {
	return &cfg.API
}

// ProvideReportGormEnginePort adapta *pgorm.Repository a report.GormEnginePort.
func ProvideReportGormEnginePort(r *pgorm.Repository) report.GormEnginePort {
	return r
}

// ProvideReportGinEnginePort adapta *pgin.Server a report.GinEnginePort.
func ProvideReportGinEnginePort(s *pgin.Server) report.GinEnginePort {
	return s
}

// ProvideReportMiddlewaresEnginePort adapta *mwr.Middlewares a report.MiddlewaresEnginePort.
func ProvideReportMiddlewaresEnginePort(m *mwr.Middlewares) report.MiddlewaresEnginePort {
	return m
}

// ReportSet expone todos los providers necesarios para Report.
var ReportSet = wire.NewSet(
	ProvideReportRepository,
	ProvideReportRepositoryPort,
	ProvideReportUseCases,
	ProvideReportUseCasesPort,
	ProvideReportHandler,
	ProvideReportConfigAPI,
	ProvideReportGormEnginePort,
	ProvideReportGinEnginePort,
	ProvideReportMiddlewaresEnginePort,
)
