package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

// Customer es el DTO para un customer específico.
type Customer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ToDomain convierte el DTO Customer a la entidad de dominio.
func (c Customer) ToDomain() *domain.Customer {
	return &domain.Customer{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
	}
}

// FromDomain convierte una entidad de dominio a su DTO.
func FromDomain(d domain.Customer) *Customer {
	return &Customer{
		ID:   d.ID,
		Name: d.Name,
		Type: d.Type,
	}
}
