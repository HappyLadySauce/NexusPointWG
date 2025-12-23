package user

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/internal/controller/user"
)

func init() {
	userController := user.NewUserController(router.StoreIns)
	router.V1().POST("/users", userController.CreateUser)
}
