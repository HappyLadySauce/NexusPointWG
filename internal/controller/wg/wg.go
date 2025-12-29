package wg

import (
	srv "github.com/HappyLadySauce/NexusPointWG/internal/service"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

// WGController handles WireGuard peer/config APIs.
type WGController struct {
	srv srv.Service
}

func NewWGController(store store.Factory) *WGController {
	return &WGController{
		srv: srv.NewService(store),
	}
}
