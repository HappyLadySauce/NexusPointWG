package v1

// RegisterRequest represents a user registration request.
// swagger:model
type RegisterRequest struct {
	// Username is the unique username for the user (3-32 characters)
	Username string `json:"username" binding:"required,min=3,max=32"`
	// Email is the user's email address (max 20 characters, must be valid email format)
	Email string `json:"email" binding:"required,email,max=20"`
	// Password is the user's password (8-32 characters, will be hashed and not returned in response)
	Password string `json:"password" binding:"required,min=8,max=32"`
}

// RegisterResponse represents a user registration response.
// swagger:model
type RegisterResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}