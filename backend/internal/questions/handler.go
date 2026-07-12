package questions

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" || h.service == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "q is required"})
		return
	}
	question, elapsed, err := h.service.Search(c.Request.Context(), query)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": "QUESTION_NOT_FOUND", "message": "question not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "question search failed"})
		return
	}
	answer := ""
	for index, item := range question.Answers {
		if index > 0 {
			answer += "###"
		}
		answer += item.Text
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{
		"question": question.Question, "answer": answer, "type": question.Type, "is_ai": false, "search_time": elapsed.Microseconds(), "sources": []string{question.Source},
	}})
}

func (h *Handler) Import(c *gin.Context) {
	var request BatchImportInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "items is required"})
		return
	}
	created, duplicates, err := h.service.Import(c.Request.Context(), request.Items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "IMPORT_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "imported", "data": gin.H{"created": created, "duplicates": duplicates}})
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("currentUser")
		user, ok := value.(auth.User)
		if !exists || !ok || user.Role != auth.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "administrator permission required"})
			return
		}
		c.Next()
	}
}
