package wg

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/wireguard"
)

// RegisterRoutes registers WireGuard management routes.
// This function must be called after router.Init() to ensure router.StoreIns is initialized.
func RegisterRoutes() {
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

	// Server configuration management routes (admin only, enforced in controller)
	authed.GET("/wg/server-config", wgController.GetServerConfig)
	authed.PUT("/wg/server-config", wgController.UpdateServerConfig)

	// Batch operations routes
	authed.POST("/wg/ip-pools/batch", wgController.BatchCreateIPPools)
	authed.PUT("/wg/ip-pools/batch", wgController.BatchUpdateIPPools)
	authed.DELETE("/wg/ip-pools/batch", wgController.BatchDeleteIPPools)
	authed.POST("/wg/peers/batch", wgController.BatchCreateWGPeers)
	authed.PUT("/wg/peers/batch", wgController.BatchUpdateWGPeers)
	authed.DELETE("/wg/peers/batch", wgController.BatchDeleteWGPeers)
}
