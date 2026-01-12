package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
)

// LeaseType DTO for LeaseType entity
type LeaseType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (l LeaseType) ToDomain() *domain.LeaseType {
	return &domain.LeaseType{
		ID:   l.ID,
		Name: l.Name,
	}
}

func FromDomain(d domain.LeaseType) *LeaseType {
	return &LeaseType{
		ID:   d.ID,
		Name: d.Name,
	}
}
