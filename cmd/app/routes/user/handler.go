package user

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/user"
)

func init() {
	userController := user.NewUserController(router.StoreIns)
	// 用户注册路由不需要认证
	router.V1().POST("/users", userController.CreateUser)
}
