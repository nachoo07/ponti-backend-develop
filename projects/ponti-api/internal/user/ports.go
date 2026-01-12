package user

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/usecases/domain"
)

type UseCases interface {
	CreateUser(context.Context, *domain.User) (string, error)
	GetUser(context.Context, string) (*domain.User, error)
	DeleteUser(context.Context, string, bool) error
	ListUsers(context.Context) ([]domain.User, error)
	UpdateUser(context.Context, *domain.User) error
	FollowUser(context.Context, string, string) (string, error)
	GetFolloweeUsers(context.Context, string) ([]string, error)
	GetFollowerUsers(context.Context, string) ([]string, error)
}

type Repository interface {
	CreateUser(context.Context, *domain.User) (string, error)
	UpdateUser(context.Context, *domain.User) error
	GetUser(context.Context, string) (*domain.User, error)
	DeleteUser(context.Context, string, bool) error
	ListUsers(context.Context) ([]domain.User, error)
	FollowUser(context.Context, string, string) (string, error)
	GetFolloweeUsers(context.Context, string) ([]string, error)
	GetFollowerUsers(context.Context, string) ([]string, error)
	FollowExists(context.Context, string, string) (bool, error)
}

// Gomock
// mockgen -source=ports.go -destination=./mocks/mock_user.go -package=mocks
