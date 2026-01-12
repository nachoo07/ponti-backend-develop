package domain

import "time"

type Investor struct {
	ID               int64     // Primary key (auto-increment)
	Name             string    // Investor's name or legal business name
	FieldID          int64     // Foreign key referencing the linked field
	Contributions    float64   // Total contribution amount (DECIMAL(10,2))
	ContributionDate time.Time // Date when the contribution was made
	Percentage       int       // Percentage of the investment in the project
}
