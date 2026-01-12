package dto

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"

type LaborCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewLaborCategoriesListResponse(categories []domain.LaborCategory) []LaborCategory {
	listedCategories := make([]LaborCategory, len(categories))

	for i, lc := range categories {
		listedCategories[i] = LaborCategory{
			ID:   lc.ID,
			Name: lc.Name,
		}
	}
	return listedCategories
}
