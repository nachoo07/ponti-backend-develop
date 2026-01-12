package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

// ListedCustomer es el DTO ligero para listados: sólo ID y Name.
type ListedCustomer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// PageInfo contiene metadata de paginación.
type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

// ListCustomersResponse es la respuesta paginada.
type ListCustomersResponse struct {
	Data     []ListedCustomer `json:"data"`
	PageInfo PageInfo         `json:"page_info"`
}

// NewListCustomersResponse construye la respuesta paginada.
func NewListCustomersResponse(
	items []domain.ListedCustomer,
	page, perPage int,
	total int64,
) ListCustomersResponse {
	out := make([]ListedCustomer, len(items))
	for i, c := range items {
		out[i] = ListedCustomer{
			ID:   c.ID,
			Name: c.Name,
		}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))
	return ListCustomersResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
