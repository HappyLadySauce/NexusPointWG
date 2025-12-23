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
	// User contains the user information
	User UserInfo `json:"user"`
}

// UserInfo contains basic user information.
// swagger:model
type UserInfo struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`
	// Username is the user's username
	Username string `json:"username"`
	// Email is the user's email address
	Email string `json:"email"`
	// Status is the user's status (active/inactive)
	Status string `json:"status"`
}
