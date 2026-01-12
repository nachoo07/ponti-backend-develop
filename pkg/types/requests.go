package pkgtypes

type LoginCredentials struct {
	Username string `json:"username,omitempty" binding:"omitempty"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Password string `json:"password" binding:"required"`
}
