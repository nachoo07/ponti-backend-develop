package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	projectmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Manager struct {
	ID       int64                `gorm:"primaryKey;autoIncrement"`
	Name     string               `gorm:"type:varchar(255);not null;unique"`
	Projects []projectmod.Project `gorm:"many2many:project_managers;"`
	sharedmodels.Base
}

// Model → Domain
func (m Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   m.ID,
		Name: m.Name,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

// Domain → Model
func FromDomain(d *domain.Manager) *Manager {
	m := &Manager{
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
	if d.ID > 0 {
		m.ID = d.ID
	}
	return m
}
