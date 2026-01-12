package pkgtypes

type LoginCredentials struct {
	Username string `json:"username,omitempty" binding:"omitempty"`    // Opcional
	Email    string `json:"email,omitempty" binding:"omitempty,email"` // Opcional
	Password string `json:"password" binding:"required"`               // Requerido
}
