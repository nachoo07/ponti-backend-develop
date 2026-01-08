package dto

type CreateInvestorResponse struct {
	Message    string `json:"message"`
	InvestorID int64  `json:"investor_id"`
}
