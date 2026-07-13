package ai

import (
	"errors"
	"net/http"
	"strconv"

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
	id, err := h.service.CreateModel(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "AI_MODEL_CREATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": gin.H{"id": id}})
}

func (h *Handler) ListModels(c *gin.Context) {
	items, err := h.service.store.ListAdminModels(c.Request.Context())
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

func (h *Handler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid AI model id"})
		return
	}
	var request UpdateModelInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid AI model payload"})
		return
	}
	if err := h.service.UpdateModel(c.Request.Context(), id, request); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "AI_MODEL_NOT_FOUND", "message": "AI model not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "AI_MODEL_UPDATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated"})
}

func (h *Handler) UpdateModelStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid AI model id"})
		return
	}
	var request struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "status is required"})
		return
	}
	if err := h.service.UpdateModelStatus(c.Request.Context(), id, request.Status); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "AI_MODEL_NOT_FOUND", "message": "AI model not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_STATUS", "message": "invalid AI model status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated"})
}

func (h *Handler) ListAnswers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := -1
	if value := c.Query("status"); value != "" {
		status, _ = strconv.Atoi(value)
	}
	data, err := h.service.ListAnswers(c.Request.Context(), c.Query("search"), c.Query("provider"), c.Query("model"), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load AI answers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

func (h *Handler) GetAnswer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid AI answer id"})
		return
	}
	item, err := h.service.GetAnswer(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": "AI_ANSWER_NOT_FOUND", "message": "AI answer not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load AI answer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
}

func (h *Handler) UpdateAnswerStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid AI answer id"})
		return
	}
	var request struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "status is required"})
		return
	}
	if err := h.service.UpdateAnswerStatus(c.Request.Context(), id, request.Status); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "AI_ANSWER_NOT_FOUND", "message": "AI answer not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_STATUS", "message": "invalid AI answer status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated"})
}
