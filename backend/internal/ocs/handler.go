package ocs

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
)

type Handler struct {
	baseURL string
	store   *Store
}

func NewHandler(baseURL string, store *Store) *Handler {
	return &Handler{baseURL: strings.TrimRight(baseURL, "/"), store: store}
}

func (h *Handler) CreateSource(c *gin.Context) {
	var input SourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid OCS source payload"})
		return
	}
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "OCS service is unavailable"})
		return
	}
	service := NewService(h.store)
	item, err := service.CreateSource(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "OCS_SOURCE_CREATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": item})
}

func (h *Handler) ListSources(c *gin.Context) {
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "OCS service is unavailable"})
		return
	}
	items, err := h.store.ListSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load OCS sources"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items})
}

func (h *Handler) UpdateSource(c *gin.Context) {
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "OCS service is unavailable"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid OCS source id"})
		return
	}
	var input SourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "invalid OCS source payload"})
		return
	}
	if err := NewService(h.store).UpdateSource(c.Request.Context(), id, input); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "OCS_SOURCE_NOT_FOUND", "message": "OCS source not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "OCS_SOURCE_UPDATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated"})
}

func (h *Handler) UpdateSourceStatus(c *gin.Context) {
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "OCS service is unavailable"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid OCS source id"})
		return
	}
	var request struct {
		Status *int `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "status is required"})
		return
	}
	if err := NewService(h.store).UpdateSourceStatus(c.Request.Context(), id, *request.Status); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "OCS_SOURCE_NOT_FOUND", "message": "OCS source not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_STATUS", "message": "invalid OCS source status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated"})
}

func (h *Handler) Config(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "API_KEY_REQUIRED", "message": "pass the full API key to generate an OCS config"})
		return
	}
	config := []map[string]any{{
		"name":        "TIKU-ZONG",
		"homepage":    h.baseURL,
		"url":         h.baseURL + "/api/ocs/search",
		"method":      "get",
		"type":        "GM_xmlhttpRequest",
		"contentType": "json",
		"data":        map[string]string{"key": key, "q": "${title}", "type": "${type}", "options": "${options}"},
		"handler":     "return (res)=> res.code === 1 ? [res.q,res.data] : undefined",
	}}
	encoded, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to generate OCS config"})
		return
	}
	if h.store != nil {
		// The generated response must contain the user's key for OCS to work,
		// but the persisted template must never contain that key in plaintext.
		persisted := []map[string]any{{
			"name":        "TIKU-ZONG",
			"homepage":    h.baseURL,
			"url":         h.baseURL + "/api/ocs/search",
			"method":      "get",
			"type":        "GM_xmlhttpRequest",
			"contentType": "json",
			"data":        map[string]string{"key": "${api_key}", "q": "${title}", "type": "${type}", "options": "${options}"},
			"handler":     "return (res)=> res.code === 1 ? [res.q,res.data] : undefined",
		}}
		if safeTemplate, marshalErr := json.Marshal(persisted); marshalErr == nil {
			_ = h.store.SaveConfig(c.Request.Context(), user.ID, string(safeTemplate))
		}
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", encoded)
}

func currentUser(c *gin.Context) (auth.User, bool) {
	value, exists := c.Get("currentUser")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
		return auth.User{}, false
	}
	user, ok := value.(auth.User)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
	}
	return user, ok
}
