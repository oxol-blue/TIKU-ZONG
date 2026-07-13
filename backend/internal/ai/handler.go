package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreateProvider(c *gin.Context) {
	var request CreateProviderInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid AI provider payload"})
		return
	}
	id, err := h.service.store.CreateProvider(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "AI_PROVIDER_CREATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": gin.H{"id": id}})
}

func (h *Handler) CreateModel(c *gin.Context) {
	var request CreateModelInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid AI model payload"})
		return
	}
	id, err := h.service.store.CreateModel(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "AI_MODEL_CREATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": gin.H{"id": id}})
}

func (h *Handler) ListModels(c *gin.Context) {
	items, err := h.service.store.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load AI models"})
		return
	}
	for index := range items {
		items[index].KeyConfigured = items[index].EncryptedKey != ""
		items[index].EncryptedKey = ""
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items})
}
