package user

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/user"
)

// RegisterRoutes registers user management routes.
// This function must be called after router.Init() to ensure router.StoreIns is initialized.
func RegisterRoutes() {
	userController := user.NewUserController(router.StoreIns)
	// 用户注册路由不需要认证
	router.V1().POST("/users", userController.CreateUser)

	// 需要认证的用户资源路由
	authed := router.Authed()
	authed.GET("/users", userController.ListUsers)
	authed.GET("/users/:username", userController.GetUserInfo)
	authed.PUT("/users/:username", userController.UpdateUserInfo)
	authed.DELETE("/users/:username", userController.DeleteUser)
	authed.POST("/users/:username/password", userController.ChangePassword)
}
