package v1

// CreateUserRequest represents a user creation request.
// swagger:model
type CreateUserRequest struct {
	// Username is the unique username for the user (3-32 characters, URL-safe, no Chinese)
	Username string `json:"username" binding:"required,min=3,max=32,urlsafe,nochinese"`
	// Nickname is the user's display name (3-32 characters). If not provided, will use username.
	Nickname string `json:"nickname" binding:"omitempty,min=3,max=32"`
	// Avatar is the user's avatar URL (must be a valid URL, max 255 characters)
	Avatar string `json:"avatar" binding:"omitempty,url,max=255"`
	// Email is the user's email address (must be valid email format and use common email provider, max 255 characters)
	Email string `json:"email" binding:"required,email,emaildomain,max=255"`
	// Password is the user's password (8-32 characters, will be hashed and not returned in response)
	Password string `json:"password" binding:"required,min=8,max=32"`
	// Role is the user role (user/admin). Only available for authenticated admin users. If not provided, defaults to "user".
	Role *string `json:"role,omitempty" binding:"omitempty,oneof=user admin"`
	// Status is the user status (active/inactive/deleted). Only available for authenticated admin users. If not provided, defaults to "active".
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive deleted"`
}

// UpdateUserRequest represents a user update request.
//
// Notes:
// - For normal users, only `username`, `nickname`, `avatar`, and `email` will be applied.
// - For admins, `password`, `status`, and `role` can also be updated.
// swagger:model
type UpdateUserRequest struct {
	// Username is the unique username for the user (3-32 characters, URL-safe, no Chinese)
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=32,urlsafe,nochinese"`
	// Nickname is the user's display name (3-32 characters)
	Nickname *string `json:"nickname,omitempty" binding:"omitempty,min=3,max=32"`
	// Avatar is the user's avatar URL (must be a valid URL, max 255 characters)
	Avatar *string `json:"avatar,omitempty" binding:"omitempty,url,max=255"`
	// Email is the user's email address (must use common email provider domain, max 255 characters)
	Email *string `json:"email,omitempty" binding:"omitempty,email,emaildomain,max=255"`
	// Password is the user's password (8-32 characters, will be hashed)
	Password *string `json:"password,omitempty" binding:"omitempty,min=8,max=32"`
	// Status is the user status (active/inactive/deleted)
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive deleted"`
	// Role is the user role (user/admin)
	Role *string `json:"role,omitempty" binding:"omitempty,oneof=user admin"`
}

// ChangePwdRequest represents a change password request.
// swagger:model
type ChangePwdRequest struct {
	// OldPassword is the user's old password (8-32 characters, will be hashed)
	OldPassword string `json:"oldPassword" binding:"required,min=8,max=32"`
	// NewPassword is the user's new password (8-32 characters, will be hashed)
	NewPassword string `json:"newPassword" binding:"required,min=8,max=32"`
}

// UserResponse represents a user response.
// swagger:model
type UserResponse struct {
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	PeerCount int64  `json:"peer_count"`
}

// UserListResponse represents a paginated list of users.
// swagger:model“
type UserListResponse struct {
	Total int64          `json:"total"`
	Items []UserResponse `json:"items"`
}
