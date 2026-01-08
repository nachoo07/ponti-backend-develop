package dto

import (
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type CreateLaborRequest struct {
	Name           string          `json:"name"`
	ContractorName string          `json:"contractor_name"`
	Price          decimal.Decimal `json:"price"`
	CategoryId     int64           `json:"category_id"`
	CreatedBy      int64           `json:"created_by"`
}

type CreateLaborsResponse struct {
	Labors  []CreateLabor `json:"labors_ids"`
	Message string        `json:"message"`
}

type CreateLabor struct {
	LaborName   string `json:"labor_name"`
	LaborID     int64  `json:"labor_id"`
	IsSaved     bool   `json:"is_saved"`
	ErrorDetail string `json:"error_detail"`
}

type LaborList struct {
	Labors []CreateLaborRequest `json:"labors"`
}

func (l CreateLaborRequest) ToDomain(projectId int64, userId int64) *domain.Labor {
	return &domain.Labor{
		Name:           l.Name,
		ContractorName: l.ContractorName,
		Price:          l.Price,
		ProjectId:      projectId,
		CategoryId:     l.CategoryId,
		Base: shareddomain.Base{
			CreatedBy: &userId,
			UpdatedBy: &userId,
		},
	}
}

func CreateLaborRequestFromDomain(d domain.Labor) *CreateLaborRequest {
	return &CreateLaborRequest{
		Name:           d.Name,
		ContractorName: d.ContractorName,
		Price:          d.Price,
	}
}
