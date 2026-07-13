package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/totp"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

type credentialsRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var request credentialsRequest
	if !bindJSON(c, &request) {
		return
	}
	if h.service == nil {
		serviceUnavailable(c)
		return
	}
	session, err := h.service.Register(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "registered", "data": session})
}

func (h *Handler) Login(c *gin.Context) {
	var request credentialsRequest
	if !bindJSON(c, &request) {
		return
	}
	if h.service == nil {
		serviceUnavailable(c)
		return
	}
	session, err := h.service.Login(c.Request.Context(), request.Email, request.Password, request.CaptchaID, request.CaptchaCode)
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": session})
}

func (h *Handler) Captcha(c *gin.Context) {
	if h.service == nil || h.service.captcha == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "captcha service is unavailable"})
		return
	}
	id, image, err := h.service.captcha.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CAPTCHA_FAILED", "message": "failed to generate captcha"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"captchaId": id, "image": image}})
}

func (h *Handler) Refresh(c *gin.Context) {
	var request refreshRequest
	if !bindJSON(c, &request) {
		return
	}
	if h.service == nil {
		serviceUnavailable(c)
		return
	}
	session, err := h.service.Refresh(c.Request.Context(), request.RefreshToken)
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": session})
}

func (h *Handler) Me(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": user})
}

func (h *Handler) GetAPIKey(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	view, err := h.service.GetAPIKey(c.Request.Context(), user.ID)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": "API_KEY_NOT_FOUND", "message": "API key not created"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": view})
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	plain, view, err := h.service.CreateAPIKey(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"code": "API_KEY_EXISTS", "message": "user already has an API key"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": gin.H{"key": plain, "info": view}})
}

func (h *Handler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.service == nil {
			serviceUnavailable(c)
			return
		}
		user, err := h.service.Authenticate(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
			return
		}
		c.Set("currentUser", user)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("currentUser")
		user, ok := value.(User)
		if !exists || !ok || user.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "permission denied"})
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc { return RequireRole(RoleAdmin) }

func RequireAdminWithTOTP(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("currentUser")
		user, ok := value.(User)
		if !exists || !ok || user.Role != RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "permission denied"})
			return
		}
		if secret != "" && !totp.Verify(secret, c.GetHeader("X-Admin-TOTP"), time.Now().UTC()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "ADMIN_TOTP_REQUIRED", "message": "valid administrator TOTP is required"})
			return
		}
		c.Next()
	}
}

// RequireBearerOrAPIKey accepts a JWT Authorization header or the public key query parameter.
func (h *Handler) RequireBearerOrAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.service == nil {
			serviceUnavailable(c)
			return
		}
		if authorization := c.GetHeader("Authorization"); authorization != "" {
			user, err := h.service.Authenticate(authorization)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
				return
			}
			c.Set("currentUser", user)
			c.Set("authMethod", "jwt")
			c.Next()
			return
		}
		user, keyID, err := h.service.AuthenticateAPIKey(c.Request.Context(), c.Query("key"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "INVALID_API_KEY", "message": "valid API key is required"})
			return
		}
		c.Set("currentUser", user)
		c.Set("apiKeyID", keyID)
		c.Set("authMethod", "api_key")
		c.Next()
	}
}

func currentUser(c *gin.Context) (User, bool) {
	value, exists := c.Get("currentUser")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
		return User{}, false
	}
	user, ok := value.(User)
	return user, ok
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid request body"})
		return false
	}
	return true
}

func handleAuthError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "authentication failed"
	switch {
	case errors.Is(err, ErrInvalidInput):
		status, code, message = http.StatusBadRequest, "INVALID_INPUT", "email or password is invalid"
	case errors.Is(err, ErrInvalidCredentials):
		status, code, message = http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect"
	case errors.Is(err, ErrAccountLocked):
		status, code, message = http.StatusLocked, "ACCOUNT_LOCKED", "account is temporarily locked"
	case errors.Is(err, ErrAccountDisabled):
		status, code, message = http.StatusForbidden, "ACCOUNT_DISABLED", "account is disabled"
	case errors.Is(err, ErrCaptchaRequired):
		status, code, message = http.StatusBadRequest, "CAPTCHA_REQUIRED", "captcha is required or invalid"
	}
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message})
}

func serviceUnavailable(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "database service is unavailable"})
}
