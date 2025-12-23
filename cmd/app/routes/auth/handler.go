package auth

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/auth"
)

func init() {
	authController := auth.NewAuthController(router.StoreIns)
	// 登录路由不需要认证
	router.V1().POST("/login", authController.Login)
}
