package dto

type CreateUser struct {
	User
}

// Response
type CreateUserResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}
