package v1

type User struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email,max=20"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}
