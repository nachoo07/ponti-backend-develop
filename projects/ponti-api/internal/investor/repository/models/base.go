package models

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
)

type Investor struct {
	ID               int64     `gorm:"primaryKey"`
	Name             string    `gorm:"type:varchar(255);not null"`
	FieldID          int64     `gorm:"not null"`
	Contributions    float64   `gorm:"type:decimal(10,2);not null"`
	ContributionDate time.Time `gorm:"type:date;not null"`
}

func (i Investor) ToDomain() *domain.Investor {
	return &domain.Investor{
		ID:               i.ID,
		Name:             i.Name,
		FieldID:          i.FieldID,
		Contributions:    i.Contributions,
		ContributionDate: i.ContributionDate,
	}
}

func FromDomain(d *domain.Investor) *Investor {
	return &Investor{
		ID:               d.ID,
		Name:             d.Name,
		FieldID:          d.FieldID,
		Contributions:    d.Contributions,
		ContributionDate: d.ContributionDate,
	}
}
