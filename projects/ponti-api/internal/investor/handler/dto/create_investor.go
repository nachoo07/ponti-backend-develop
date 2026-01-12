package dto

// CreateInvestor is the DTO for the create request of an investor.
// It embeds the base Investor DTO.
type CreateInvestor struct {
	Investor
}

type CreateInvestorResponse struct {
	Message    string `json:"message"`
	InvestorID int64  `json:"investor_id"`
}
