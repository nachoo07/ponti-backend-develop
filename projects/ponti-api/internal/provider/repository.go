// File: ./repository.go

package provider

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
	"gorm.io/gorm"
)

// GormEnginePort exposes the required DB interface.
type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	var providers []models.Provider
	if err := r.db.Client().WithContext(ctx).Find(&providers).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list providers", err)
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}
