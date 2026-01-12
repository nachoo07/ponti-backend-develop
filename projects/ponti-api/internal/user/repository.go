package user

import (
	"context"
	"fmt"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	"github.com/google/uuid"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository crea una nueva implementación del repositorio de usuarios.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

// CreateUser persists a new user into the database.
func (r *repository) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is nil")
	}

	// Convertir la entidad del dominio al modelo de persistencia.
	model, err := models.FromDomain(user)
	if err != nil {
		return "", fmt.Errorf("error converting domain user to model: %w", err)
	}
	model.ID = uuid.New().String()

	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return "", fmt.Errorf("error creating user in database: %w", err)
	}

	return model.ID, nil
}

// ListUsers retrieves all users from the database.
func (r *repository) ListUsers(ctx context.Context) ([]domain.User, error) {
	var modelsList []models.User
	if err := r.db.Client().WithContext(ctx).Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	users := make([]domain.User, 0, len(modelsList))
	for _, m := range modelsList {
		user, err := m.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("error converting model to domain: %w", err)
		}
		users = append(users, *user)
	}
	return users, nil
}

// GetUser retrieves a user by its ID.
func (r *repository) GetUser(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}

	var model models.User
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, fmt.Errorf("error retrieving user with id %s: %w", id, err)
	}

	user, err := model.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("error converting model to domain: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user in the database.
func (r *repository) UpdateUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	model, err := models.FromDomain(user)
	if err != nil {
		return fmt.Errorf("error converting domain user to model: %w", err)
	}

	if err := r.db.Client().WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("error updating user with id %s: %w", user.ID, err)
	}
	return nil
}

// DeleteUser deletes a user by its ID.
func (r *repository) DeleteUser(ctx context.Context, id string, hardDelete bool) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}
	db := r.db.Client().WithContext(ctx)
	if hardDelete {
		db = db.Unscoped()
	}
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("error deleting user with id %s: %w", id, err)
	}
	return nil
}

// FollowUser creates a follow relationship between two users.
// El repository se limita a persistir la relación sin aplicar reglas de negocio.
func (r *repository) FollowUser(ctx context.Context, followerID, followeeID string) (string, error) {
	if followerID == "" || followeeID == "" {
		return "", fmt.Errorf("followerID or followeeID is empty")
	}

	followModel := models.Follow{
		ID:         uuid.New().String(),
		FollowerID: followerID,
		FolloweeID: followeeID,
	}

	if err := r.db.Client().WithContext(ctx).Create(&followModel).Error; err != nil {
		return "", fmt.Errorf("error creating follow relationship: %w", err)
	}

	return followModel.ID, nil
}

// GetFolloweeUsers retrieves the list of followee user IDs for a given follower.
func (r *repository) GetFolloweeUsers(ctx context.Context, followerID string) ([]string, error) {
	if followerID == "" {
		return nil, fmt.Errorf("followerID is empty")
	}

	var follows []models.Follow
	if err := r.db.Client().WithContext(ctx).
		Where("follower_id = ?", followerID).
		Find(&follows).Error; err != nil {
		return nil, fmt.Errorf("error retrieving follow relationships: %w", err)
	}

	followeeIDs := make([]string, 0, len(follows))
	for _, f := range follows {
		followeeIDs = append(followeeIDs, f.FolloweeID)
	}

	return followeeIDs, nil
}

// FollowExists checks whether a follow relationship between follower and followee already exists.
func (r *repository) FollowExists(ctx context.Context, followerID, followeeID string) (bool, error) {
	var count int64
	err := r.db.Client().
		WithContext(ctx).
		Model(&models.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("error checking follow existence: %w", err)
	}
	return count > 0, nil
}

func (r *repository) GetFollowerUsers(ctx context.Context, followeeID string) ([]string, error) {
	if followeeID == "" {
		return nil, fmt.Errorf("followeeID is empty")
	}

	var follows []models.Follow
	if err := r.db.Client().WithContext(ctx).
		Where("followee_id = ?", followeeID).
		Find(&follows).Error; err != nil {
		return nil, fmt.Errorf("error retrieving follower relationships: %w", err)
	}

	followerIDs := make([]string, 0, len(follows))
	for _, f := range follows {
		followerIDs = append(followerIDs, f.FollowerID)
	}

	return followerIDs, nil
}
