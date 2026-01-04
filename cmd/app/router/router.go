package router

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/internal/store/sqlite"
	"github.com/HappyLadySauce/NexusPointWG/pkg/environment"

	_ "github.com/HappyLadySauce/NexusPointWG/api/swagger/docs"
	_ "github.com/HappyLadySauce/NexusPointWG/pkg/utils/validator"
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

	// setup static file serving for frontend
	setupStaticFiles(router)
}

// setupStaticFiles configures static file serving for the frontend SPA.
// It checks multiple possible locations for the frontend build output:
// 1. /app/ui (Docker container path)
// 2. _output/dist (local development path)
// 3. ui/dist (alternative local path)
// If found, serves static files and handles SPA routing by returning index.html for non-API routes.
func setupStaticFiles(r *gin.Engine) {
	// Possible static file directories (in order of preference)
	staticDirs := []string{
		"/app/ui",                              // Docker container path
		"_output/dist",                         // Local build output
		"ui/dist",                              // Alternative local path
		filepath.Join("..", "_output", "dist"), // Relative path from binary location
	}

	var staticDir string
	for _, dir := range staticDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Check if index.html exists in this directory
			indexPath := filepath.Join(dir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				staticDir = dir
				klog.V(2).InfoS("Found static files directory", "path", staticDir)
				break
			}
		}
	}

	if staticDir == "" {
		klog.V(2).InfoS("Static files directory not found, skipping static file serving")
		return
	}

	// Serve static files
	r.Static("/static", filepath.Join(staticDir, "static"))
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))
	r.StaticFile("/vite.svg", filepath.Join(staticDir, "vite.svg"))

	// SPA fallback: serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// If the path has an extension, try to serve it as a static file
		if filepath.Ext(path) != "" {
			filePath := filepath.Join(staticDir, path)
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}
		}

		// For all other routes (including SPA routes), serve index.html
		c.File(filepath.Join(staticDir, "index.html"))
	})
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
