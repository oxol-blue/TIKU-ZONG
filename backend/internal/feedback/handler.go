package feedback

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid feedback payload"})
		return
	}
	value, exists := c.Get("currentUser")
	user, ok := value.(auth.User)
	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
		return
	}
	if err := h.service.Create(c.Request.Context(), user.ID, input); err != nil {
		status := http.StatusBadRequest
		code := "FEEDBACK_FAILED"
		if errors.Is(err, ErrDuplicate) {
			status = http.StatusConflict
			code = "FEEDBACK_EXISTS"
		}
		c.JSON(status, gin.H{"code": code, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "feedback recorded"})
}
