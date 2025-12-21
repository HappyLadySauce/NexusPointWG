package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/pprof"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
)

func SetupMiddlewares(router *gin.Engine) {
	
	// install cors middleware
	router.Use(middleware.Cors())

	// install pprof handler
	pprof.Register(router)

	// install metrics handler
	prometheus := ginprometheus.NewPrometheus("gin")
	prometheus.Use(router)
}