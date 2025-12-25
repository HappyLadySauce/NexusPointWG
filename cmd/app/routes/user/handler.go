package user

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/user"
)

func init() {
	userController := user.NewUserController(router.StoreIns)
	// 用户注册路由不需要认证
	router.V1().POST("/users", userController.RegisterUser)

	// 需要认证的用户资源路由
	authed := router.Authed()
	authed.GET("/users/:id", userController.GetUserInfo)
	authed.PUT("/users/:id", userController.UpdateUserInfo)
	authed.DELETE("/users/:id", userController.DeleteUser)
}
