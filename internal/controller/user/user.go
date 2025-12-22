package user

import (
	srv "github.com/HappyLadySauce/NexusPointWG/internal/service"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

// UserController create a user handler used to handle request for user resource.
type UserController struct {
	srv srv.Service
}

// NewUserController creates a user handler.
func NewUserController(store store.Factory) *UserController {
	return &UserController{
		srv: srv.NewService(store),
	}
}
