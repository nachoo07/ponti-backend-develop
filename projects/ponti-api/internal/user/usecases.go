package user

import (
	"context"
	"fmt"

	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/usecases/domain"
)

type useCases struct {
	repository Repository
}

// NewUseCases crea una nueva instancia de useCases.
func NewUseCases(rp Repository) UseCases {
	return &useCases{
		repository: rp,
	}
}

// CreateUser creates a new user by hashing the password (business logic) and storing the user using the repository.
func (u *useCases) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is nil")
	}

	// La transformación de la contraseña se hace en el use case (regla de negocio)
	hashedPassword, err := utils.HashPassword(user.Credentials.Password, 12)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	user.Credentials.Password = hashedPassword

	newUserID, err := u.repository.CreateUser(ctx, user)
	if err != nil {
		return "", fmt.Errorf("error creating user: %w", err)
	}

	return newUserID, nil
}

// ListUsers retrieves a list of all users.
func (u *useCases) ListUsers(ctx context.Context) ([]domain.User, error) {
	users, err := u.repository.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	return users, nil
}

// GetUser retrieves a user by its ID.
func (u *useCases) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is empty")
	}

	user, err := u.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user with ID %s: %w", userID, err)
	}
	return user, nil
}

// DeleteUser deletes a user by its ID.
func (u *useCases) DeleteUser(ctx context.Context, id string, hardDelete bool) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}

	if err := u.repository.DeleteUser(ctx, id, hardDelete); err != nil {
		return fmt.Errorf("error deleting user with ID %s: %w", id, err)
	}
	return nil
}

// UpdateUser updates the user's information.
func (u *useCases) UpdateUser(ctx context.Context, updatedUser *domain.User) error {
	if updatedUser == nil {
		return fmt.Errorf("updatedUser is nil")
	}

	if err := u.repository.UpdateUser(ctx, updatedUser); err != nil {
		return fmt.Errorf("error updating user with ID %s: %w", updatedUser.ID, err)
	}
	return nil
}

// FollowUser creates a follow relationship between two users.
func (u *useCases) FollowUser(ctx context.Context, followerID, followeeID string) (string, error) {
	if followerID == "" || followeeID == "" {
		return "", fmt.Errorf("followerID or followeeID is empty")
	}

	// Verifica que el seguidor exista.
	if _, err := u.repository.GetUser(ctx, followerID); err != nil {
		return "", fmt.Errorf("follower with ID %s does not exist: %w", followerID, err)
	}

	// Verifica que el seguido exista.
	if _, err := u.repository.GetUser(ctx, followeeID); err != nil {
		return "", fmt.Errorf("followee with ID %s does not exist: %w", followeeID, err)
	}

	exists, err := u.repository.FollowExists(ctx, followerID, followeeID)
	if err != nil {
		return "", fmt.Errorf("error checking follow relationship: %w", err)
	}
	if exists {
		return "", fmt.Errorf("follower %s already follows user %s", followerID, followeeID)
	}

	relationID, err := u.repository.FollowUser(ctx, followerID, followeeID)
	if err != nil {
		return "", fmt.Errorf("error following user: %w", err)
	}
	return relationID, nil
}

// GetFolloweeUsers retrieves the list of user IDs that the given follower is following.
func (u *useCases) GetFolloweeUsers(ctx context.Context, followerID string) ([]string, error) {
	if followerID == "" {
		return nil, fmt.Errorf("followerID is empty")
	}

	followees, err := u.repository.GetFolloweeUsers(ctx, followerID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving followees for user %s: %w", followerID, err)
	}
	return followees, nil
}

// GetFollowerUsers retrieves the list of user IDs that are following the given user (followee).
func (u *useCases) GetFollowerUsers(ctx context.Context, followeeID string) ([]string, error) {
	if followeeID == "" {
		return nil, fmt.Errorf("followeeID is empty")
	}

	followers, err := u.repository.GetFollowerUsers(ctx, followeeID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving followers for user %s: %w", followeeID, err)
	}
	return followers, nil
}
