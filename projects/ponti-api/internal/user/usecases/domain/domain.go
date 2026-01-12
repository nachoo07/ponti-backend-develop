package domain

import (
	"time"
)

type UserType string

const (
	UserTypePerson UserType = "person"
)

type User struct {
	ID             string
	Credentials    Credentials
	PersonID       string
	UserType       UserType
	Roles          []Role
	LoggedAt       time.Time
	EmailValidated bool
}

type Credentials struct {
	Email    string
	Password string
}

type Role struct {
	Name        string
	Permissions []Permission
}

type Permission struct {
	Name        string
	Description string
}

type Follow struct {
	FollowerID string // seguidor
	FolloweeID string // seguido
}
