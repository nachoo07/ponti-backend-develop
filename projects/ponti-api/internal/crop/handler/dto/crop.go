package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
)

// Crop represents a crop for data transfer.
type Crop struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ToDomain converts the DTO Crop to the domain entity.
func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:   c.ID,
		Name: c.Name,
	}
}

// FromDomain converts a domain Crop to the DTO.
func FromDomain(d domain.Crop) *Crop {
	return &Crop{
		ID:   d.ID,
		Name: d.Name,
	}
}

func NewListCropsResponse(crops []domain.Crop) *ListCropsResponse {
	var convertedCrops []Crop
	for _, crop := range crops {
		convertedCrops = append(convertedCrops, *FromDomain(crop))
	}
	return &ListCropsResponse{
		Crops: convertedCrops,
	}
}

// ListCropsResponse represents a list of crops for data transfer.

type ListCropsResponse struct {
	Crops []Crop `json:"data"`
}
