package wireguard

import (
	srv "github.com/HappyLadySauce/NexusPointWG/internal/service"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

// WGController creates a WireGuard handler used to handle requests for WireGuard resources.
type WGController struct {
	srv srv.Service
}

// NewWGController creates a WireGuard handler.
func NewWGController(store store.Factory) *WGController {
	return &WGController{
		srv: srv.NewService(store),
	}
}
