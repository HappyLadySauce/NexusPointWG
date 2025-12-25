package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/internal/store/sqlite"
	"github.com/HappyLadySauce/NexusPointWG/pkg/environment"

	_ "github.com/HappyLadySauce/NexusPointWG/api/swagger/docs"
)

var (
	router *gin.Engine
	v1     *gin.RouterGroup
	authed *gin.RouterGroup

	StoreIns store.Factory
)

func init() {
	var err error
	StoreIns, err = sqlite.GetSqliteFactoryOr(nil)
	if err != nil {
		// In init(), we can't return error, so we panic if store initialization fails
		// This will prevent the application from starting with an invalid store
		// Use %+v to show the full error chain including the original error
		klog.Fatalf("Failed to initialize store in router init: %+v", err)
	}

	if !environment.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()

	// setup middlewares
	SetupMiddlewares(router)

	// setup routes
	_ = router.SetTrustedProxies(nil)
	v1 = router.Group("/api/v1")

	authed = v1.Group("/")
	authed.Use(middleware.JWTAuth(StoreIns))

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

func Authed() *gin.RouterGroup {
	return authed
}

// Router returns the main Gin engine instance.
func Router() *gin.Engine {
	return router
}
