package wireguard

import (
	"github.com/HappyLadySauce/NexusPointWG/internal/service"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

type WireGuardController struct {
	srv service.Service
}

func NewWireGuardController(storeIns store.Factory) *WireGuardController {
	return &WireGuardController{srv: service.NewService(storeIns)}
}
