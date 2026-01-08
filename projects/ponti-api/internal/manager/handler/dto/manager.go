package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

// Manager es el DTO para un manager específico.
type Manager struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ToDomain convierte el DTO Manager a la entidad de dominio.
func (c Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
	}
}

// FromDomain convierte una entidad de dominio a su DTO.
func FromDomain(d domain.Manager) *Manager {
	return &Manager{
		ID:   d.ID,
		Name: d.Name,
		Type: d.Type,
	}
}
