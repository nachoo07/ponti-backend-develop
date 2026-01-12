package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

type Customer struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
	Type string `gorm:"type:varchar(100);not null"`
}

func (c Customer) ToDomain() *domain.Customer {
	return &domain.Customer{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
	}
}

func FromDomain(d *domain.Customer) *Customer {
	return &Customer{
		ID:   d.ID,
		Name: d.Name,
		Type: d.Type,
	}
}
