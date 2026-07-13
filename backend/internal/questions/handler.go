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
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/ocs"
)

type Handler struct {
	service *Service
	logger  *calls.Store
	billing *billing.Service
	ai      *ai.Service
	ocs     *ocs.Service
}

func NewHandler(service *Service, logger *calls.Store, billingService *billing.Service, aiService *ai.Service, ocsServices ...*ocs.Service) *Handler {
	var ocsService *ocs.Service
	if len(ocsServices) > 0 {
		ocsService = ocsServices[0]
	}
	return &Handler{service: service, logger: logger, billing: billingService, ai: aiService, ocs: ocsService}
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
		h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusBadRequest, "INVALID_QUERY", started)
		if isOCS == true {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "q is required"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_QUERY", "message": "q is required"})
		}
		return
	}
	options := splitOptions(c.Query("options"))
	question, elapsed, similarity, err := h.service.SearchWithScore(c.Request.Context(), query, options)
	if errors.Is(err, ErrNotFound) {
		questionType := c.Query("type")
		if h.ocs != nil {
			if external, externalErr := h.ocs.Search(c.Request.Context(), query, questionType, options); externalErr == nil {
				if h.billing != nil {
					packageID, _ := strconv.ParseUint(c.Query("package_id"), 10, 64)
					if _, consumeErr := h.billing.Consume(c.Request.Context(), current.ID, packageID, billing.UsageQuestions, requestID, "/api/v1/search", 1); consumeErr != nil {
						h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusPaymentRequired, "NO_QUOTA", started)
						c.JSON(http.StatusPaymentRequired, gin.H{"code": "NO_QUOTA", "message": "an available package is required"})
						return
					}
				}
				h.log(c, requestID, current.ID, keyID, query, true, "ocs", http.StatusOK, "", started)
				h.recordSearch(c, current.ID, requestID, external.Question, questionType, external.Answer, external.Source, false, started)
				if isOCS == true {
					c.JSON(http.StatusOK, gin.H{"code": 1, "q": external.Question, "data": external.Answer})
				} else {
					c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"request_id": requestID, "question": external.Question, "answer": external.Answer, "type": questionType, "is_ai": false, "search_time": external.Elapsed.Microseconds(), "sources": []string{external.Source}}})
				}
				return
			}
		}
		if h.ai != nil {
			packageID, _ := strconv.ParseUint(c.Query("package_id"), 10, 64)
			if h.billing != nil {
				available, quotaErr := h.billing.HasAIQuota(c.Request.Context(), current.ID, packageID)
				if quotaErr != nil || !available {
					h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusPaymentRequired, "NO_AI_QUOTA", started)
					c.JSON(http.StatusPaymentRequired, gin.H{"code": "NO_AI_QUOTA", "message": "an available AI package quota is required"})
					return
				}
			}
			aiAnswer, aiErr := h.ai.Solve(c.Request.Context(), query, questionType, options)
			if aiErr == nil {
				if h.billing != nil {
					if _, consumeErr := h.billing.Consume(c.Request.Context(), current.ID, packageID, billing.UsageAI, requestID, "/api/v1/search", aiAnswer.ChargeCount); consumeErr != nil {
						h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusPaymentRequired, "NO_AI_QUOTA", started)
						c.JSON(http.StatusPaymentRequired, gin.H{"code": "NO_AI_QUOTA", "message": "an available AI package quota is required"})
						return
					}
				}
				h.logAI(c, requestID, current.ID, keyID, query, true, http.StatusOK, "", started)
				h.recordSearch(c, current.ID, requestID, query, questionType, aiAnswer.Text, aiAnswer.Provider+"/"+aiAnswer.Model, true, started)
				if isOCS == true {
					c.JSON(http.StatusOK, gin.H{"code": 1, "q": query, "data": aiAnswer.Text})
				} else {
					c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"request_id": requestID, "question": query, "answer": aiAnswer.Text, "type": questionType, "is_ai": true, "search_time": time.Since(started).Microseconds(), "sources": []string{aiAnswer.Provider + "/" + aiAnswer.Model}}})
				}
				return
			}
		}
		h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusNotFound, "QUESTION_NOT_FOUND", started)
		if isOCS == true {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "question not found"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"code": "QUESTION_NOT_FOUND", "message": "question not found"})
		}
		return
	}
	if err != nil {
		h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusInternalServerError, "INTERNAL_ERROR", started)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "question search failed"})
		return
	}
	if h.billing != nil {
		packageID, _ := strconv.ParseUint(c.Query("package_id"), 10, 64)
		if _, err := h.billing.Consume(c.Request.Context(), current.ID, packageID, billing.UsageQuestions, requestID, "/api/v1/search", 1); err != nil {
			h.log(c, requestID, current.ID, keyID, query, false, "", http.StatusPaymentRequired, "NO_QUOTA", started)
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
	h.log(c, requestID, current.ID, keyID, query, true, "local", http.StatusOK, "", started)
	h.recordSearch(c, current.ID, requestID, question.Question, question.Type, answer, question.Source, false, started)
	if isOCS == true {
		c.JSON(http.StatusOK, gin.H{"code": 1, "q": question.Question, "data": answer})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{
		"request_id": requestID, "question": question.Question, "answer": answer, "type": question.Type, "is_ai": false, "similarity": similarity, "search_time": elapsed.Microseconds(), "sources": []string{question.Source},
	}})
}

func splitOptions(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return strings.Split(value, "\n")
}

func (h *Handler) OCSSearch(c *gin.Context) {
	c.Set("ocsMode", true)
	h.Search(c)
}

func (h *Handler) log(c *gin.Context, requestID string, userID, keyID uint64, question string, success bool, sourceKind string, status int, errorCode string, started time.Time) {
	if h.logger == nil {
		return
	}
	endpoint := "/api/v1/search"
	if isOCS, _ := c.Get("ocsMode"); isOCS == true {
		endpoint = "/api/ocs/search"
	}
	_ = h.logger.Log(c.Request.Context(), calls.Log{RequestID: requestID, UserID: userID, APIKeyID: keyID, Endpoint: endpoint, Question: question, Success: success, SourceKind: sourceKind, HTTPStatus: status, ErrorCode: errorCode, Elapsed: time.Since(started)})
}

func (h *Handler) logAI(c *gin.Context, requestID string, userID, keyID uint64, question string, success bool, status int, errorCode string, started time.Time) {
	if h.logger == nil {
		return
	}
	endpoint := "/api/v1/search"
	if isOCS, _ := c.Get("ocsMode"); isOCS == true {
		endpoint = "/api/ocs/search"
	}
	_ = h.logger.Log(c.Request.Context(), calls.Log{RequestID: requestID, UserID: userID, APIKeyID: keyID, Endpoint: endpoint, Question: question, Success: success, IsAI: true, SourceKind: "ai", HTTPStatus: status, ErrorCode: errorCode, Elapsed: time.Since(started)})
}

func (h *Handler) recordSearch(c *gin.Context, userID uint64, requestID, question, questionType, answer, source string, isAI bool, started time.Time) {
	if h.logger == nil {
		return
	}
	_ = h.logger.RecordSearch(c.Request.Context(), calls.SearchHistory{
		UserID: userID, RequestID: requestID, Question: question, Type: questionType,
		Answer: answer, Source: source, IsAI: isAI, Elapsed: time.Since(started),
	})
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

// ImportFile accepts a CSV or XLSX workbook with Chinese or English column names.
func (h *Handler) ImportFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE", "message": "file is required"})
		return
	}
	items, report, err := parseImportFile(fileHeader)
	if err != nil {
		if report.Total > 0 && report.Invalid > 0 {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "no valid rows", "data": report})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE", "message": err.Error(), "data": report})
		return
	}
	created, duplicates, err := h.service.Import(c.Request.Context(), items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "IMPORT_FAILED", "message": err.Error(), "data": report})
		return
	}
	report.Created, report.Duplicates = created, duplicates
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "imported", "data": report})
}

func (h *Handler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := -1
	if value := c.Query("status"); value != "" {
		status, _ = strconv.Atoi(value)
	}
	data, err := h.service.ListAdmin(c.Request.Context(), c.Query("search"), c.Query("type"), c.Query("subject"), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load questions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

func (h *Handler) AdminDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid question id"})
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": "QUESTION_NOT_FOUND", "message": "question not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load question"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
}

func (h *Handler) AdminUpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "invalid question id"})
		return
	}
	var request struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "status is required"})
		return
	}
	if err := h.service.UpdateStatus(c.Request.Context(), id, request.Status); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "QUESTION_NOT_FOUND", "message": "question not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_STATUS", "message": "invalid question status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated"})
}
