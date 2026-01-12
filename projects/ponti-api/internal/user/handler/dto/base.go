package dto

import (
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/user/usecases/domain"
)

type User struct {
	UserType       string                 `json:"user_type"`
	EmailValidated bool                   `json:"email_validated"`
	PersonID       string                 `json:"person_id"`
	Credentials    types.LoginCredentials `json:"credentials"`
	Roles          []Role                 `json:"roles"`
}

type Role struct {
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Follow struct {
	FollowerID string `json:"follower_id" binding:"required"`
	FolloweeID string `json:"followee_id" binding:"required"`
}

// Mappers
func (dto *User) ToDomain() *domain.User {
	user := &domain.User{
		UserType:       domain.UserType(dto.UserType),
		EmailValidated: dto.EmailValidated,
		PersonID:       dto.PersonID, // Asignamos directamente como string
		Credentials: domain.Credentials{
			Email:    dto.Credentials.Email,    // Asegúrate de que types.LoginCredentials tenga estos campos
			Password: dto.Credentials.Password, // Si no, ajusta según corresponda
		},
		Roles: convertRoles(dto.Roles),
	}

	return user
}

// Función auxiliar para convertir los roles del DTO al dominio
func convertRoles(dtoRoles []Role) []domain.Role {
	roles := make([]domain.Role, len(dtoRoles))
	for i, roleDTO := range dtoRoles {
		permissions := make([]domain.Permission, len(roleDTO.Permissions))
		for j, permDTO := range roleDTO.Permissions {
			permissions[j] = domain.Permission{
				Name:        permDTO.Name,
				Description: permDTO.Description,
			}
		}
		roles[i] = domain.Role{
			Name:        roleDTO.Name,
			Permissions: permissions,
		}
	}
	return roles
}
