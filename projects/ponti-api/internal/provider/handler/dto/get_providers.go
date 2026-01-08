package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
)

type Provider struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewGetProvidersResponse(providersDomain []*domain.Provider) []Provider {
	var providers []Provider

	for _, provider := range providersDomain {
		providers = append(
			providers,
			Provider{
				ID:   provider.ID,
				Name: provider.Name,
			},
		)
	}

	return providers
}
