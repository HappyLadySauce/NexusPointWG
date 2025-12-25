package v1

// LoginRequest represents a login request.
// swagger:model
type LoginRequest struct {
	// Username is the user's username
	Username string `json:"username" binding:"required"`
	// Password is the user's password
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response.
// swagger:model
type LoginResponse struct {
	// Token is the JWT token for authentication
	Token string `json:"token"`
}

