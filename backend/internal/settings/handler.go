package settings

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Public(c *gin.Context) {
	config, err := h.service.Public(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load public settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": config})
}

func (h *Handler) Admin(c *gin.Context) {
	config, err := h.service.Admin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load system settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": config})
}

func (h *Handler) Update(c *gin.Context) {
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid system settings payload"})
		return
	}
	config, err := h.service.Update(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		if errors.Is(err, ErrInvalidInput) {
			status, code = http.StatusBadRequest, "INVALID_SETTINGS"
		}
		c.JSON(status, gin.H{"code": code, "message": "failed to save system settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated", "data": config})
}
