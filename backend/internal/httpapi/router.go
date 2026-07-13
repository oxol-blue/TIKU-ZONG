package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ai"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/announcement"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/audit"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/billing"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/calls"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/feedback"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ocs"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/payment"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/questions"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/security"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/settings"
)

// NewRouter builds the public HTTP router for the API service.
func NewRouter(cfg config.Config, authService *auth.Service, questionService *questions.Service, billingService *billing.Service, callLogger *calls.Store, aiService *ai.Service, ocsStore *ocs.Store, services ...any) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(security.NewMiddleware(security.Config{RateLimitPerMinute: cfg.APIRateLimitPerMinute, Blacklist: cfg.IPBlacklist, Whitelist: cfg.IPWhitelist, RedisAddr: cfg.RedisAddr}))
	var paymentService *payment.Service
	var feedbackService *feedback.Service
	var ocsService *ocs.Service
	var announcementService *announcement.Service
	var settingsService *settings.Service
	var databaseService *sql.DB
	for _, service := range services {
		switch value := service.(type) {
		case *payment.Service:
			paymentService = value
		case *feedback.Service:
			feedbackService = value
		case *ocs.Service:
			ocsService = value
		case *sql.DB:
			databaseService = value
		case *announcement.Service:
			announcementService = value
		case *settings.Service:
			settingsService = value
		}
	}
	var auditStore *audit.Store
	if databaseService != nil {
		auditStore = audit.NewStore(databaseService)
	}
	_ = router.SetTrustedProxies(nil)
	var requestCount uint64
	router.Use(gin.Logger(), gin.Recovery(), func(c *gin.Context) {
		atomic.AddUint64(&requestCount, 1)
		c.Next()
	}, corsMiddleware(), adminAuditMiddleware(auditStore))

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
	router.GET("/readyz", func(c *gin.Context) {
		if databaseService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"code": "NOT_READY", "message": "database is not configured"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		if err := databaseService.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"code": "NOT_READY", "message": "database is unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ready"})
	})
	router.GET("/metrics", func(c *gin.Context) {
		stats := aiService.QueueStats()
		c.Header("Content-Type", "text/plain; version=0.0.4")
		c.String(http.StatusOK, "tiku_http_requests_total %d\ntiku_ai_queue_depth %d\ntiku_ai_queue_capacity %d\ntiku_ai_queue_workers %d\n", atomic.LoadUint64(&requestCount), stats.Depth, stats.Capacity, stats.Workers)
	})

	api := router.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
	})

	authHandler := auth.NewHandler(authService)
	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", registrationGuard(settingsService), authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	authRoutes.GET("/captcha", authHandler.Captcha)
	authRoutes.POST("/refresh", authHandler.Refresh)

	protected := api.Group("")
	protected.Use(authHandler.RequireAuth())
	protected.GET("/me", authHandler.Me)
	protected.POST("/password/change", authHandler.ChangePassword)
	protected.GET("/api-key", authHandler.GetAPIKey)
	protected.POST("/api-key", authHandler.CreateAPIKey)
	protected.POST("/api-key/rotate", authHandler.RotateAPIKey)
	questionHandler := questions.NewHandler(questionService, callLogger, billingService, aiService, ocsService)
	aiHandler := ai.NewHandler(aiService)
	ocsHandler := ocs.NewHandler(cfg.PublicBaseURL, ocsStore)
	billingHandler := billing.NewHandler(billingService)
	callHandler := calls.NewHandler(callLogger)
	questionAuth := api.Group("")
	questionAuth.Use(authHandler.RequireBearerOrAPIKey())
	questionAuth.GET("/search", questionHandler.Search)
	adminOnly := auth.RequireAdminWithTOTP(cfg.AdminTOTPSecret)
	if announcementService != nil {
		announcementHandler := announcement.NewHandler(announcementService)
		protected.GET("/announcements", announcementHandler.List)
		protected.GET("/admin/announcements", adminOnly, announcementHandler.AdminList)
		protected.POST("/admin/announcements", adminOnly, announcementHandler.Create)
		protected.PUT("/admin/announcements/:id", adminOnly, announcementHandler.Update)
		protected.PATCH("/admin/announcements/:id/status", adminOnly, announcementHandler.UpdateStatus)
	}
	protected.POST("/admin/questions/import", adminOnly, questionHandler.Import)
	protected.POST("/admin/questions/import/file", adminOnly, questionHandler.ImportFile)
	protected.GET("/admin/questions", adminOnly, questionHandler.AdminList)
	protected.GET("/admin/questions/:id", adminOnly, questionHandler.AdminDetail)
	protected.PATCH("/admin/questions/:id/status", adminOnly, questionHandler.AdminUpdateStatus)
	protected.GET("/admin/users", adminOnly, authHandler.AdminUsers)
	protected.POST("/admin/invites", adminOnly, authHandler.AdminCreateInvite)
	protected.GET("/admin/invites", adminOnly, authHandler.AdminListInvites)
	protected.PATCH("/admin/users/:id/status", adminOnly, authHandler.AdminUpdateUserStatus)
	protected.PATCH("/admin/users/:id/role", adminOnly, authHandler.AdminUpdateUserRole)
	protected.GET("/packages/my", billingHandler.MyPackages)
	protected.GET("/packages", billingHandler.AvailablePackages)
	protected.POST("/admin/packages", adminOnly, billingHandler.CreatePackage)
	protected.GET("/admin/packages", adminOnly, billingHandler.AdminListPackages)
	protected.PUT("/admin/packages/:id", adminOnly, billingHandler.AdminUpdatePackage)
	protected.PATCH("/admin/packages/:id/status", adminOnly, billingHandler.AdminUpdatePackageStatus)
	protected.POST("/admin/coupons", adminOnly, billingHandler.CreateCoupon)
	protected.GET("/admin/coupons", adminOnly, billingHandler.ListCoupons)
	protected.PATCH("/admin/coupons/:id/status", adminOnly, billingHandler.AdminUpdateCouponStatus)
	protected.POST("/admin/packages/:id/grant/:userId", adminOnly, billingHandler.GrantPackage)
	protected.GET("/admin/calls", adminOnly, callHandler.Recent)
	protected.GET("/calls/my", callHandler.Mine)
	protected.GET("/admin/dashboard", adminOnly, callHandler.Dashboard)
	if settingsService != nil {
		settingsHandler := settings.NewHandler(settingsService)
		api.GET("/settings/public", settingsHandler.Public)
		protected.GET("/admin/settings", adminOnly, settingsHandler.Admin)
		protected.PUT("/admin/settings", adminOnly, settingsHandler.Update)
	}
	if auditStore != nil {
		auditHandler := audit.NewHandler(auditStore)
		protected.GET("/admin/audit-logs", adminOnly, auditHandler.List)
	}
	protected.POST("/admin/ai/providers", adminOnly, aiHandler.CreateProvider)
	protected.POST("/admin/ai/models", adminOnly, aiHandler.CreateModel)
	protected.GET("/admin/ai/models", adminOnly, aiHandler.ListModels)
	protected.GET("/admin/ai/answers", adminOnly, aiHandler.ListAnswers)
	protected.GET("/admin/ai/answers/:id", adminOnly, aiHandler.GetAnswer)
	protected.PATCH("/admin/ai/answers/:id/status", adminOnly, aiHandler.UpdateAnswerStatus)
	protected.POST("/admin/ocs/sources", adminOnly, ocsHandler.CreateSource)
	protected.GET("/admin/ocs/sources", adminOnly, ocsHandler.ListSources)
	ocsRoutes := router.Group("/api/ocs")
	ocsRoutes.Use(authHandler.RequireBearerOrAPIKey())
	ocsRoutes.GET("/config", ocsHandler.Config)
	ocsRoutes.GET("/search", questionHandler.OCSSearch)
	if feedbackService != nil {
		feedbackHandler := feedback.NewHandler(feedbackService)
		protected.POST("/feedback", feedbackHandler.Create)
		protected.GET("/feedback/my", feedbackHandler.Mine)
		protected.GET("/admin/feedback", adminOnly, feedbackHandler.AdminList)
	}
	if paymentService != nil {
		paymentHandler := payment.NewHandler(paymentService)
		protected.POST("/orders", paymentHandler.CreateOrder)
		protected.GET("/orders/my", paymentHandler.MyOrders)
		protected.GET("/admin/orders", adminOnly, paymentHandler.AdminOrders)
		protected.POST("/admin/payment/gateways", adminOnly, paymentHandler.ConfigureGateway)
		protected.GET("/admin/payment/gateways", adminOnly, paymentHandler.Gateway)
		protected.POST("/admin/orders/close-expired", adminOnly, paymentHandler.CloseExpired)
		protected.POST("/admin/orders/:orderNo/refund", adminOnly, paymentHandler.Refund)
		protected.GET("/admin/orders/:orderNo/refunds", adminOnly, paymentHandler.Refunds)
		protected.GET("/admin/orders/reconciliation", adminOnly, paymentHandler.Reconciliation)
		publicPayment := router.Group("/api/payment")
		publicPayment.GET("/notify/:provider", paymentHandler.Notify)
		publicPayment.POST("/notify/:provider", paymentHandler.Notify)
	}

	return router
}

func registrationGuard(service *settings.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil {
			c.Next()
			return
		}
		enabled, err := service.RegistrationEnabled(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": "SETTINGS_UNAVAILABLE", "message": "registration is temporarily unavailable"})
			return
		}
		if !enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "REGISTRATION_CLOSED", "message": "registration is closed"})
			return
		}
		c.Next()
	}
}

// adminAuditMiddleware records successful state changes made through admin APIs.
// It deliberately stores no request body, so passwords, provider keys, and payment
// credentials cannot be copied into the operation log.
func adminAuditMiddleware(store *audit.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if store == nil || !strings.HasPrefix(c.Request.URL.Path, "/api/v1/admin/") {
			return
		}
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			return
		}
		if status := c.Writer.Status(); status < http.StatusOK || status >= http.StatusBadRequest {
			return
		}
		value, exists := c.Get("currentUser")
		user, ok := value.(auth.User)
		if !exists || !ok || user.Role != auth.RoleAdmin {
			return
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		resource := strings.TrimPrefix(path, "/api/v1/admin/")
		_ = store.Create(c.Request.Context(), audit.Entry{
			AdminID: user.ID, AdminEmail: user.Email, Action: c.Request.Method,
			Resource: resource, RequestPath: path, IPAddress: c.ClientIP(), HTTPStatus: c.Writer.Status(),
		})
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Admin-TOTP")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
