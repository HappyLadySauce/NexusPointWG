package wg

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/wg"
)

func init() {
	wgController := wg.NewWGController(router.StoreIns)

	authed := router.Authed()

	// Admin peer management
	authed.GET("/wg/peers", wgController.ListPeers)
	authed.POST("/wg/peers", wgController.CreatePeer)
	authed.GET("/wg/peers/:id", wgController.GetPeer)
	authed.PUT("/wg/peers/:id", wgController.UpdatePeer)
	authed.DELETE("/wg/peers/:id", wgController.DeletePeer)

	// User configs
	authed.GET("/wg/configs", wgController.ListMyConfigs)
	authed.GET("/wg/configs/:id/download", wgController.DownloadConfig)
	authed.POST("/wg/configs/:id/rotate", wgController.RotateConfig)
	authed.POST("/wg/configs/:id/revoke", wgController.RevokeConfig)
}
