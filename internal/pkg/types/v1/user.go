package v1


type User struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email,max=20"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}