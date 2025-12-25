package wireguard

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/wireguard"
)

func init() {
	wgController := wireguard.NewWireGuardController(router.StoreIns)
	authed := router.Authed()

	authed.GET("/wg/peers", wgController.ListPeers)
	authed.POST("/wg/peers", wgController.CreatePeer)
	authed.GET("/wg/peers/:id", wgController.GetPeer)
	authed.PUT("/wg/peers/:id", wgController.UpdatePeer)
	authed.DELETE("/wg/peers/:id", wgController.DeletePeer)
	authed.GET("/wg/peers/:id/config", wgController.DownloadPeerConfig)
}
