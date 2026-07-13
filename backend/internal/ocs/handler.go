package ocs

import (
	"encoding/json"
	"net/http"
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
