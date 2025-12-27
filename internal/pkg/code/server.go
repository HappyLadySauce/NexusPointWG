package code

// Server: server-related errors.
// Code must start with 1xxxxx.
const (
	// ErrUserAlreadyExist - 400: User already exists.
	ErrUserAlreadyExist int = iota + 110001

	// ErrEmailAlreadyExist - 400: Email already exists.
	ErrEmailAlreadyExist
	
	// ErrUserNotFound - 404: User not found.
	ErrUserNotFound

	// ErrUserNotActive - 403: User account is not active.
	ErrUserNotActive
)
