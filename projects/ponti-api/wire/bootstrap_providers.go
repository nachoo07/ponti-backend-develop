package wire

import (
	"fmt"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pgdb "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/postgresql/pgxpool"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	ssmtp "github.com/alphacodinggroup/ponti-backend/pkg/notification/smtp"
)

func ProvideGormRepository() (gorm.Repository, error) {
	repo, err := gorm.Bootstrap("", "", "", "", "", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gorm: %w", err)
	}

	return repo, nil
}

func ProvideGinServer() (ginsrv.Server, error) {
	isTest := false
	server, err := ginsrv.Bootstrap("", "", isTest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gin server: %w", err)
	}
	return server, nil
}

func ProvidePostgresRepository() (pgdb.Repository, error) {
	repo, err := pgdb.Bootstrap("", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap PostgreSQL repository: %w", err)
	}
	return repo, nil
}

func ProvideSmtpService() (ssmtp.Service, error) {
	ssmtp, err := ssmtp.Bootstrap("", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SMTP service: %w", err)
	}

	return ssmtp, nil
}
