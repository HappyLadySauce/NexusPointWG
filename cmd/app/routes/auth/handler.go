package auth

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/auth"
)

// RegisterRoutes registers authentication routes.
// This function must be called after router.Init() to ensure router.StoreIns is initialized.
func RegisterRoutes() {
	authController := auth.NewAuthController(router.StoreIns)
	// 登录路由不需要认证
	router.V1().POST("/login", authController.Login)
}
