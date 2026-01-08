package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type ListedProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type ListProjectsResponse struct {
	Data     []ListedProject `json:"data"`
	PageInfo PageInfo        `json:"page_info"`
}

func NewListProjectsResponse(
	items []domain.ListedProject,
	page, perPage int,
	total int64,
) ListProjectsResponse {
	out := make([]ListedProject, len(items))
	for i, p := range items {
		out[i] = ListedProject{ID: int64(i + 1), Name: p.Name}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))

	return ListProjectsResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}

type ProjectsResponse struct {
	Data          []Project       `json:"data"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
	PageInfo      PageInfo        `json:"page_info"`
}

// MarshalJSON aplica redondeo de 3 decimales al campo TotalHectares
func (p ProjectsResponse) MarshalJSON() ([]byte, error) {
	aux := struct {
		Data          []Project `json:"data"`
		TotalHectares string    `json:"total_hectares"`
		PageInfo      PageInfo  `json:"page_info"`
	}{
		Data:          p.Data,
		TotalHectares: p.TotalHectares.Round(3).String(),
		PageInfo:      p.PageInfo,
	}
	return json.Marshal(aux)
}

func NewProjectsResponse(
	items []domain.Project,
	page, perPage int,
	total int64,
) ProjectsResponse {
	out := make([]Project, len(items))
	for i, p := range items {
		out[i] = Project{
			ID:          p.ID,
			ProjectName: p.Name,
			Customer: Customer{
				ID:   p.Customer.ID,
				Name: p.Customer.Name,
			},
			Campaign: Campaign{
				ID:   p.Campaign.ID,
				Name: p.Campaign.Name,
			},
		}

		for _, mgr := range p.Managers {
			out[i].ProjectManagers = append(out[i].ProjectManagers,
				Manager{ID: mgr.ID, Name: mgr.Name},
			)
		}

		for _, inv := range p.Investors {
			out[i].Investors = append(out[i].Investors,
				Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage},
			)
		}

		for _, aci := range p.AdminCostInvestors {
			out[i].AdminCostInvestors = append(out[i].AdminCostInvestors,
				AdminCostInvestor{ID: aci.ID, Name: aci.Name, Percentage: aci.Percentage},
			)
		}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))

	return ProjectsResponse{
		Data: out,
		PageInfo: PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
