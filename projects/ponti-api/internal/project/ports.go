package project

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type UseCases interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	ListProjects(context.Context) ([]domain.Project, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
	ListProjectsByCustomerID(context.Context, int64) ([]domain.Project, error)
}

type Repository interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	ListProjects(context.Context) ([]domain.Project, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
	ListProjectsByCustomerID(context.Context, int64) ([]domain.Project, error)
}
