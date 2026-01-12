package models

import (
	"errors"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/usecases/domain"
)

// Mappers
func FromDomainFollow(f *domain.Follow) (*Follow, error) {
	if f == nil {
		return nil, errors.New("follow is nil")
	}

	return &Follow{
		FollowerID: f.FollowerID,
		FolloweeID: f.FolloweeID,
	}, nil
}

func (fm *Follow) ToDomainFollow() (*domain.Follow, error) {
	if fm == nil {
		return nil, errors.New("follow model is nil")
	}

	return &domain.Follow{
		FollowerID: fm.FollowerID,
		FolloweeID: fm.FolloweeID,
	}, nil
}
