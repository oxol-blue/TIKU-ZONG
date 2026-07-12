package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/questions"
)

// NewRouter builds the public HTTP router for the API service.
func NewRouter(cfg config.Config, authService *auth.Service, questionService *questions.Service) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"service": cfg.AppName,
				"env":     cfg.AppEnv,
			},
		})
	})

	api := router.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
	})

	authHandler := auth.NewHandler(authService)
	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	authRoutes.POST("/refresh", authHandler.Refresh)

	protected := api.Group("")
	protected.Use(authHandler.RequireAuth())
	protected.GET("/me", authHandler.Me)
	protected.GET("/api-key", authHandler.GetAPIKey)
	protected.POST("/api-key", authHandler.CreateAPIKey)

	questionHandler := questions.NewHandler(questionService)
	protected.GET("/search", questionHandler.Search)
	protected.POST("/admin/questions/import", questions.RequireAdmin(), questionHandler.Import)

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
