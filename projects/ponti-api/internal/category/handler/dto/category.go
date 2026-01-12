package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/usecases/domain"
)

type Category struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name"`
	TypeID int64  `json:"type_id"`
}

func (d *Category) ToDomain() *domain.Category {
	return &domain.Category{
		ID:     d.ID,
		Name:   d.Name,
		TypeID: d.TypeID,
	}
}

func FromDomain(c *domain.Category) *Category {
	return &Category{
		ID:     c.ID,
		Name:   c.Name,
		TypeID: c.TypeID,
	}
}
