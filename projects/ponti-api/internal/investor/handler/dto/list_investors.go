package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

// ListedInvestor is the lightweight DTO for list operations.
type ListedInvestor struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ListInvestorsResponse holds the list of investors.
type ListInvestorsResponse struct {
	Data []ListedInvestor `json:"data"`
}

// NewListInvestorsResponse constructs the response DTO.
func NewListInvestorsResponse(items []domain.ListedInvestor) ListInvestorsResponse {
	out := make([]ListedInvestor, len(items))
	for i, inv := range items {
		out[i] = ListedInvestor{ID: inv.ID, Name: inv.Name}
	}
	return ListInvestorsResponse{Data: out}
}
