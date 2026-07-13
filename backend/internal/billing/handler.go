package billing

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreatePackage(c *gin.Context) {
	var request CreatePackageInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid package payload"})
		return
	}
	item, err := h.service.CreatePackage(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PACKAGE", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": item})
}

func (h *Handler) GrantPackage(c *gin.Context) {
	packageID, err1 := strconv.ParseUint(c.Param("id"), 10, 64)
	userID, err2 := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid package or user id"})
		return
	}
	item, err := h.service.GrantPackage(c.Request.Context(), userID, packageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "GRANT_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "granted", "data": item})
}

func (h *Handler) MyPackages(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	items, err := h.service.ListInstances(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load packages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items})
}

func currentUser(c *gin.Context) (auth.User, bool) {
	value, exists := c.Get("currentUser")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
		return auth.User{}, false
	}
	if current, valid := value.(auth.User); valid {
		return current, true
	}
	return auth.User{}, false
}
