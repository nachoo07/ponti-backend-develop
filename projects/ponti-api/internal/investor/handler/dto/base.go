package dto

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

// Investor is the DTO for a specific investor.
type Investor struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	FieldID          int64     `json:"field_id"`
	Contributions    float64   `json:"contributions"`
	ContributionDate time.Time `json:"contribution_date"`
}

// ToDomain converts the DTO Investor to the domain entity.
func (i Investor) ToDomain() *domain.Investor {
	return &domain.Investor{
		ID:               i.ID,
		Name:             i.Name,
		FieldID:          i.FieldID,
		Contributions:    i.Contributions,
		ContributionDate: i.ContributionDate,
	}
}

// FromDomain converts a domain Investor to the DTO.
func FromDomain(d domain.Investor) *Investor {
	return &Investor{
		ID:               d.ID,
		Name:             d.Name,
		FieldID:          d.FieldID,
		Contributions:    d.Contributions,
		ContributionDate: d.ContributionDate,
	}
}
