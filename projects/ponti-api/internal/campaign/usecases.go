package campaign

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type RepositoryPort interface {
	CreateCampaign(context.Context, *domain.Campaign) (int64, error)
	ListCampaigns(context.Context, int64, string) ([]domain.Campaign, error)
	GetCampaign(context.Context, int64) (*domain.Campaign, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error) {
	return u.repo.CreateCampaign(ctx, c)
}

func (u *UseCases) ListCampaigns(ctx context.Context, customerID int64, projectName string) ([]domain.Campaign, error) {
	return u.repo.ListCampaigns(ctx, customerID, projectName)
}

func (u *UseCases) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	return u.repo.GetCampaign(ctx, id)
}
