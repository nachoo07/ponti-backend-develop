package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Investor struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Percentage int    `json:"percentage"`
	shareddomain.Base
}

type ListedInvestor struct {
	ID   int64  // Primary key
	Name string // Investor name
}
