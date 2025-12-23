package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/internal/store/sqlite"
	"github.com/HappyLadySauce/NexusPointWG/pkg/environment"
	
	_ "github.com/HappyLadySauce/NexusPointWG/api/swagger/docs"
)

var (
	router *gin.Engine
	v1     *gin.RouterGroup

	StoreIns store.Factory
)

func init() {
	StoreIns, _ = sqlite.GetSqliteFactoryOr(nil)
	
	if !environment.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()

	// setup middlewares
	SetupMiddlewares(router)

	// setup routes
	_ = router.SetTrustedProxies(nil)
	v1 = router.Group("/api/v1")

	router.GET("/livez", func(c *gin.Context) {
		c.String(200, "livez")
	})
	router.GET("/readyz", func(c *gin.Context) {
		c.String(200, "readyz")
	})

	// register swagger routes
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// V1 returns the router group for /api/v1 which for resources in control plane endpoints.
func V1() *gin.RouterGroup {
	return v1
}

// Router returns the main Gin engine instance.
func Router() *gin.Engine {
	return router
}
