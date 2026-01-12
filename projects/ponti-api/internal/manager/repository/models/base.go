package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

type Manager struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"column:name;type:varchar(100);not null"`
	Type string `gorm:"column:type;type:varchar(50);not null"`
}

func (c Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
	}
}

func FromDomain(d *domain.Manager) *Manager {
	return &Manager{
		ID:   d.ID,
		Name: d.Name,
		Type: d.Type,
	}
}
