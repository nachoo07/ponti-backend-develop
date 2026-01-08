package dto

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Labor struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	ContractorName string          `json:"contractor_name"`
	Price          decimal.Decimal `json:"price"`
	CategoryId     int64           `json:"category_id"`
}

func (l Labor) ToDomain(projectId int64, userId int64) *domain.Labor {
	return &domain.Labor{
		ID:             l.ID,
		Name:           l.Name,
		ContractorName: l.ContractorName,
		Price:          l.Price,
		ProjectId:      projectId,
		CategoryId:     l.CategoryId,
		Base: shareddomain.Base{
			UpdatedBy: &userId,
		},
	}
}

func FromDomain(d domain.Labor) *Labor {
	return &Labor{
		ID:             d.ID,
		Name:           d.Name,
		ContractorName: d.ContractorName,
		Price:          d.Price,
	}

}
