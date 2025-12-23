package v1

// User represents a user creation request and response.
// swagger:model
type User struct {
	// Username is the unique username for the user (3-32 characters)
	Username string `json:"username" binding:"required,min=3,max=32"`
	// Email is the user's email address (max 20 characters, must be valid email format)
	Email string `json:"email" binding:"required,email,max=20"`
	// Password is the user's password (8-32 characters, will be hashed and not returned in response)
	Password string `json:"password" binding:"required,min=8,max=32"`
}
