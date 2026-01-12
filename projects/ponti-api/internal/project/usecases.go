package project

import (
	"context"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domainField "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type RepositoryPort interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	ListProjects(context.Context, int, int) ([]domain.ListedProject, int64, error)
	GetProjects(context.Context, string, int64, int64, int, int) ([]domain.Project, decimal.Decimal, int64, error)
	ListProjectsByCustomerID(context.Context, int64, int, int) ([]domain.ListedProject, int64, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	GetProjectByNameAndCampaignID(context.Context, string, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
	RestoreProject(context.Context, int64) error
	GetFieldsByProjectID(context.Context, int64) ([]domainField.Field, error)
}

type WordsSuggesterPort interface {
	Suggest(ctx context.Context, prefix string, page, perPage int) ([]domain.ListedProject, int64, error)
	Close() error
	Health(context.Context) error
}

type UseCases struct {
	repo           RepositoryPort
	wordsSuggester WordsSuggesterPort
}

func NewUseCases(
	rp RepositoryPort,
	sg WordsSuggesterPort,
) *UseCases {
	return &UseCases{
		wordsSuggester: sg,
		repo:           rp,
	}
}

func (u *UseCases) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	exist, err := u.repo.GetProjectByNameAndCampaignID(ctx, p.Name, p.Campaign.ID)
	if err != nil {
		return 0, err
	}

	if exist != nil {
		return 0, types.NewError(types.ErrConflict, "project already exists", nil)
	}

	return u.repo.CreateProject(ctx, p)
}

func (u *UseCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	return u.repo.GetProject(ctx, id)
}

func (u *UseCases) GetProjects(ctx context.Context, name string, customerID int64, campaignID int64, page, perPage int) ([]domain.Project, decimal.Decimal, int64, error) {
	return u.repo.GetProjects(ctx, name, customerID, campaignID, page, perPage)
}

func (u *UseCases) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	return u.repo.ListProjects(ctx, page, perPage)
}

func (u *UseCases) GetFieldsByProjectID(ctx context.Context, projectID int64) ([]domainField.Field, error) {
	return u.repo.GetFieldsByProjectID(ctx, projectID)
}

func (u *UseCases) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	return u.repo.ListProjectsByCustomerID(ctx, customerID, page, perPage)
}

func (u *UseCases) UpdateProject(ctx context.Context, p *domain.Project) error {
	exist, err := u.repo.GetProjectByNameAndCampaignID(ctx, p.Name, p.Campaign.ID)
	if err != nil {
		return err
	}

	if exist != nil && exist.ID != p.ID {
		return types.NewError(types.ErrConflict, "project already exists", nil)
	}
	return u.repo.UpdateProject(ctx, p)
}

func (u *UseCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}

func (u *UseCases) RestoreProject(ctx context.Context, id int64) error {
	return u.repo.RestoreProject(ctx, id)
}

func (u *UseCases) ListProjectsByName(ctx context.Context, name string, page, perPage int) ([]domain.ListedProject, int64, error) {
	return u.wordsSuggester.Suggest(ctx, name, page, perPage)
}
