package wg

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/wireguard"
)

func init() {
	wgController := wireguard.NewWGController(router.StoreIns)

	// WireGuard routes require authentication
	authed := router.Authed()

	// Peer management routes
	authed.POST("/wg/peers", wgController.CreatePeer)
	authed.GET("/wg/peers", wgController.ListPeers)
	authed.GET("/wg/peers/:id", wgController.GetPeer)
	authed.PUT("/wg/peers/:id", wgController.UpdatePeer)
	authed.DELETE("/wg/peers/:id", wgController.DeletePeer)
	authed.GET("/wg/peers/:id/config", wgController.DownloadPeerConfig)

	// IP pool management routes (admin only, enforced in controller)
	authed.POST("/wg/ip-pools", wgController.CreateIPPool)
	authed.GET("/wg/ip-pools", wgController.ListIPPools)
	authed.PUT("/wg/ip-pools/:id", wgController.UpdateIPPool)
	authed.DELETE("/wg/ip-pools/:id", wgController.DeleteIPPool)
	authed.GET("/wg/ip-pools/:id/available-ips", wgController.GetAvailableIPs)
}
