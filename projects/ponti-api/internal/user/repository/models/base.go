package models

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/usecases/domain"
)

type User struct {
	ID             string         `gorm:"primaryKey;column:id"`
	PersonID       *string        `gorm:"column:person_id"`
	Email          string         `gorm:"column:email;unique;not null"`
	Password       string         `gorm:"column:password;not null"`
	EmailValidated bool           `gorm:"column:email_validated;default:false"`
	UserType       string         `gorm:"column:user_type;not null"`
	LoggedAt       *time.Time     `gorm:"column:logged_at"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index;column:deleted_at"`
	// Los roles se manejarán a través de una tabla de unión (no se incluyen directamente aquí)
}

type Role struct {
	ID          string         `gorm:"primaryKey;column:id"`
	Name        string         `gorm:"column:name;unique;not null"`
	Description string         `gorm:"column:description"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

type Permission struct {
	ID          string         `gorm:"primaryKey;column:id"`
	Name        string         `gorm:"column:name;unique;not null"`
	Description string         `gorm:"column:description"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

type UserRole struct {
	UserID string `gorm:"column:user_id;primaryKey"`
	RoleID string `gorm:"column:role_id;primaryKey"`
}

type RolePermission struct {
	RoleID       string `gorm:"column:role_id;primaryKey"`
	PermissionID string `gorm:"column:permission_id;primaryKey"`
}

type Follow struct {
	ID         string    `gorm:"primaryKey;column:id"`
	FollowerID string    `gorm:"column:follower_id;not null"`
	FolloweeID string    `gorm:"column:followee_id;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

// Mappers
func FromDomain(u *domain.User) (*User, error) {
	if u == nil {
		return nil, errors.New("user cannot be nil")
	}

	var personID *string
	if u.PersonID != "" {
		personID = &u.PersonID
	}

	var loggedAt *time.Time
	if !u.LoggedAt.IsZero() {
		loggedAt = &u.LoggedAt
	}

	return &User{
		ID:             u.ID,
		PersonID:       personID,
		Email:          u.Credentials.Email,
		Password:       u.Credentials.Password,
		EmailValidated: u.EmailValidated,
		UserType:       string(u.UserType),
		LoggedAt:       loggedAt,
	}, nil
}

func (um *User) ToDomain() (*domain.User, error) {
	if um == nil {
		return nil, errors.New("user model is nil")
	}

	var personID string
	if um.PersonID != nil {
		personID = *um.PersonID
	}

	var loggedAt time.Time
	if um.LoggedAt != nil {
		loggedAt = *um.LoggedAt
	}

	return &domain.User{
		ID:             um.ID,
		PersonID:       personID,
		EmailValidated: um.EmailValidated,
		Credentials: domain.Credentials{
			Email:    um.Email,
			Password: um.Password,
		},
		UserType: domain.UserType(um.UserType),
		LoggedAt: loggedAt,
		Roles:    []domain.Role{},
	}, nil
}
