package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/pprof"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/pkg/environment"
)

func SetupMiddlewares(router *gin.Engine) {
	
	// install cors middleware
	router.Use(middleware.Cors())

	// install pprof handler and metrics handler only in development mode
	if !environment.IsDev() {
		// install pprof handler
		pprof.Register(router)

		// install metrics handler
		prometheus := ginprometheus.NewPrometheus("gin")
		prometheus.Use(router)
	}
}