package questions

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ai"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/billing"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/calls"
)

type Handler struct {
	service *Service
	logger  *calls.Store
	billing *billing.Service
	ai      *ai.Service
}

func NewHandler(service *Service, logger *calls.Store, billingService *billing.Service, aiService *ai.Service) *Handler {
	return &Handler{service: service, logger: logger, billing: billingService, ai: aiService}
}

func (h *Handler) Search(c *gin.Context) {
	isOCS, _ := c.Get("ocsMode")
	query := strings.TrimSpace(c.Query("q"))
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	user, _ := c.Get("currentUser")
	current, _ := user.(auth.User)
	var keyID uint64
	if value, exists := c.Get("apiKeyID"); exists {
		keyID, _ = value.(uint64)
	}
	started := time.Now()
	if query == "" || h.service == nil {
		h.log(c, requestID, current.ID, keyID, query, false, http.StatusBadRequest, "INVALID_QUERY", started)
		if isOCS == true {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "q is required"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "q is required"})
		}
		return
	}
	question, elapsed, err := h.service.Search(c.Request.Context(), query)
	if errors.Is(err, ErrNotFound) {
		if h.ai != nil {
			questionType := c.Query("type")
			options := strings.Split(strings.TrimSpace(c.Query("options")), "\n")
			aiAnswer, aiErr := h.ai.Solve(c.Request.Context(), query, questionType, options)
			if aiErr == nil {
				if h.billing != nil {
					packageID, _ := strconv.ParseUint(c.Query("package_id"), 10, 64)
					if _, consumeErr := h.billing.Consume(c.Request.Context(), current.ID, packageID, billing.UsageAI, requestID, "/api/v1/search", aiAnswer.ChargeCount); consumeErr != nil {
						h.log(c, requestID, current.ID, keyID, query, false, http.StatusPaymentRequired, "NO_AI_QUOTA", started)
						c.JSON(http.StatusPaymentRequired, gin.H{"code": "NO_AI_QUOTA", "message": "an available AI package quota is required"})
						return
					}
				}
				h.logAI(c, requestID, current.ID, keyID, query, true, http.StatusOK, "", started)
				if isOCS == true {
					c.JSON(http.StatusOK, gin.H{"code": 1, "q": query, "data": aiAnswer.Text})
				} else {
					c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"question": query, "answer": aiAnswer.Text, "type": questionType, "is_ai": true, "search_time": time.Since(started).Microseconds(), "sources": []string{aiAnswer.Provider + "/" + aiAnswer.Model}}})
				}
				return
			}
		}
		h.log(c, requestID, current.ID, keyID, query, false, http.StatusNotFound, "QUESTION_NOT_FOUND", started)
		if isOCS == true {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "question not found"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"code": "QUESTION_NOT_FOUND", "message": "question not found"})
		}
		return
	}
	if err != nil {
		h.log(c, requestID, current.ID, keyID, query, false, http.StatusInternalServerError, "INTERNAL_ERROR", started)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "question search failed"})
		return
	}
	if h.billing != nil {
		packageID, _ := strconv.ParseUint(c.Query("package_id"), 10, 64)
		if _, err := h.billing.Consume(c.Request.Context(), current.ID, packageID, billing.UsageQuestions, requestID, "/api/v1/search", 1); err != nil {
			h.log(c, requestID, current.ID, keyID, query, false, http.StatusPaymentRequired, "NO_QUOTA", started)
			c.JSON(http.StatusPaymentRequired, gin.H{"code": "NO_QUOTA", "message": "an available package is required"})
			return
		}
	}
	answer := ""
	for index, item := range question.Answers {
		if index > 0 {
			answer += "###"
		}
		answer += item.Text
	}
	h.log(c, requestID, current.ID, keyID, query, true, http.StatusOK, "", started)
	if isOCS == true {
		c.JSON(http.StatusOK, gin.H{"code": 1, "q": question.Question, "data": answer})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{
		"question": question.Question, "answer": answer, "type": question.Type, "is_ai": false, "search_time": elapsed.Microseconds(), "sources": []string{question.Source},
	}})
}

func (h *Handler) OCSSearch(c *gin.Context) {
	c.Set("ocsMode", true)
	h.Search(c)
}

func (h *Handler) log(c *gin.Context, requestID string, userID, keyID uint64, question string, success bool, status int, errorCode string, started time.Time) {
	if h.logger == nil {
		return
	}
	_ = h.logger.Log(c.Request.Context(), calls.Log{RequestID: requestID, UserID: userID, APIKeyID: keyID, Endpoint: "/api/v1/search", Question: question, Success: success, HTTPStatus: status, ErrorCode: errorCode, Elapsed: time.Since(started)})
}

func (h *Handler) logAI(c *gin.Context, requestID string, userID, keyID uint64, question string, success bool, status int, errorCode string, started time.Time) {
	if h.logger == nil {
		return
	}
	_ = h.logger.Log(c.Request.Context(), calls.Log{RequestID: requestID, UserID: userID, APIKeyID: keyID, Endpoint: "/api/v1/search", Question: question, Success: success, IsAI: true, HTTPStatus: status, ErrorCode: errorCode, Elapsed: time.Since(started)})
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
