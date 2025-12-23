package auth

import (
	srv "github.com/HappyLadySauce/NexusPointWG/internal/service"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

// AuthController creates an auth handler used to handle authentication requests.
type AuthController struct {
	srv        srv.Service
}

// NewAuthController creates an auth handler.
func NewAuthController(store store.Factory) *AuthController {
	return &AuthController{
		srv:        srv.NewService(store),
	}
}
