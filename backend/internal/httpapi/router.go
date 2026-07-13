package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ai"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/billing"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/calls"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/feedback"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ocs"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/payment"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/questions"
)

// NewRouter builds the public HTTP router for the API service.
func NewRouter(cfg config.Config, authService *auth.Service, questionService *questions.Service, billingService *billing.Service, callLogger *calls.Store, aiService *ai.Service, ocsStore *ocs.Store, services ...any) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	var paymentService *payment.Service
	var feedbackService *feedback.Service
	for _, service := range services {
		switch value := service.(type) {
		case *payment.Service:
			paymentService = value
		case *feedback.Service:
			feedbackService = value
		}
	}
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
	authRoutes.GET("/captcha", authHandler.Captcha)
	authRoutes.POST("/refresh", authHandler.Refresh)

	protected := api.Group("")
	protected.Use(authHandler.RequireAuth())
	protected.GET("/me", authHandler.Me)
	protected.GET("/api-key", authHandler.GetAPIKey)
	protected.POST("/api-key", authHandler.CreateAPIKey)

	questionHandler := questions.NewHandler(questionService, callLogger, billingService, aiService)
	aiHandler := ai.NewHandler(aiService)
	ocsHandler := ocs.NewHandler(cfg.PublicBaseURL, ocsStore)
	billingHandler := billing.NewHandler(billingService)
	callHandler := calls.NewHandler(callLogger)
	questionAuth := api.Group("")
	questionAuth.Use(authHandler.RequireBearerOrAPIKey())
	questionAuth.GET("/search", questionHandler.Search)
	protected.POST("/admin/questions/import", auth.RequireAdmin(), questionHandler.Import)
	protected.GET("/packages/my", billingHandler.MyPackages)
	protected.GET("/packages", billingHandler.AvailablePackages)
	protected.POST("/admin/packages", auth.RequireAdmin(), billingHandler.CreatePackage)
	protected.POST("/admin/coupons", auth.RequireAdmin(), billingHandler.CreateCoupon)
	protected.GET("/admin/coupons", auth.RequireAdmin(), billingHandler.ListCoupons)
	protected.POST("/admin/packages/:id/grant/:userId", auth.RequireAdmin(), billingHandler.GrantPackage)
	protected.GET("/admin/calls", auth.RequireAdmin(), callHandler.Recent)
	protected.POST("/admin/ai/providers", auth.RequireAdmin(), aiHandler.CreateProvider)
	protected.POST("/admin/ai/models", auth.RequireAdmin(), aiHandler.CreateModel)
	protected.GET("/admin/ai/models", auth.RequireAdmin(), aiHandler.ListModels)
	ocsRoutes := router.Group("/api/ocs")
	ocsRoutes.Use(authHandler.RequireBearerOrAPIKey())
	ocsRoutes.GET("/config", ocsHandler.Config)
	ocsRoutes.GET("/search", questionHandler.OCSSearch)
	if feedbackService != nil {
		protected.POST("/feedback", feedback.NewHandler(feedbackService).Create)
	}
	if paymentService != nil {
		paymentHandler := payment.NewHandler(paymentService)
		protected.POST("/orders", paymentHandler.CreateOrder)
		protected.GET("/orders/my", paymentHandler.MyOrders)
		protected.POST("/admin/payment/gateways", auth.RequireAdmin(), paymentHandler.ConfigureGateway)
		protected.POST("/admin/orders/close-expired", auth.RequireAdmin(), paymentHandler.CloseExpired)
		protected.POST("/admin/orders/:orderNo/refund", auth.RequireAdmin(), paymentHandler.Refund)
		publicPayment := router.Group("/api/payment")
		publicPayment.GET("/notify/:provider", paymentHandler.Notify)
		publicPayment.POST("/notify/:provider", paymentHandler.Notify)
	}

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
