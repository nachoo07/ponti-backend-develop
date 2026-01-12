package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
)

type ClassType struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

func (d *ClassType) ToDomain() *domain.ClassType {
	return &domain.ClassType{
		ID:   d.ID,
		Name: d.Name,
	}
}
func FromDomain(c *domain.ClassType) *ClassType {
	return &ClassType{
		ID:   c.ID,
		Name: c.Name,
	}
}
