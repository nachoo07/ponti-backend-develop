package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type ListedLabor struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	CategoryId     int64           `json:"category_id"`
	Price          decimal.Decimal `json:"price"`
	ContractorName string          `json:"contractor_name"`
	CategoryName   string          `json:"category_name"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// MarshalJSON aplica redondeo de 2 decimales al precio (Costo u$s/ha)
func (l ListedLabor) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID             int64     `json:"id"`
		Name           string    `json:"name"`
		CategoryId     int64     `json:"category_id"`
		Price          string    `json:"price"`
		ContractorName string    `json:"contractor_name"`
		CategoryName   string    `json:"category_name"`
		UpdatedAt      time.Time `json:"updated_at"`
	}{
		ID:             l.ID,
		Name:           l.Name,
		CategoryId:     l.CategoryId,
		Price:          l.Price.StringFixed(2), // Costo u$s/ha: 2 decimales
		ContractorName: l.ContractorName,
		CategoryName:   l.CategoryName,
		UpdatedAt:      l.UpdatedAt,
	}
	return json.Marshal(aux)
}

// PageInfo contiene metadata de paginación.
type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type ListLaborResponse struct {
	Data     []ListedLabor `json:"data"`
	PageInfo PageInfo      `json:"page_info"`
}

func NewListLaborsResponse(
	items []domain.ListedLabor,
	page, perPage int,
	total int64,
) ListLaborResponse {
	out := make([]ListedLabor, len(items))
	for i, l := range items {
		out[i] = ListedLabor{
			ID:             l.ID,
			Name:           l.Name,
			CategoryId:     l.CategoryId,
			Price:          l.Price,
			ContractorName: l.ContractorName,
			CategoryName:   l.CategoryName,
			UpdatedAt:      l.UpdatedAt,
		}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))
	return ListLaborResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
