package payment

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oxol-blue/TIKU-ZONG/backend/internal/auth"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreateOrder(c *gin.Context) {
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "packageId is required"})
		return
	}
	user, ok := currentUser(c)
	if !ok || h.service == nil {
		return
	}
	order, paymentURL, err := h.service.CreateOrder(c.Request.Context(), user.ID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ORDER_CREATE_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": gin.H{"order": order, "paymentUrl": paymentURL}})
}

func (h *Handler) MyOrders(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok || h.service == nil {
		return
	}
	items, err := h.service.Store().ListOrders(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items})
}

func (h *Handler) AdminOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	data, err := h.service.Store().ListAdminOrders(c.Request.Context(), c.Query("search"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to load orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

func (h *Handler) Notify(c *gin.Context) {
	if h.service == nil {
		c.String(http.StatusServiceUnavailable, "fail")
		return
	}
	provider := c.Param("provider")
	values := url.Values{}
	for name, items := range c.Request.URL.Query() {
		for _, item := range items {
			values.Add(name, item)
		}
	}
	if c.Request.Method == http.MethodPost {
		_ = c.Request.ParseForm()
		for name, items := range c.Request.PostForm {
			for _, item := range items {
				values.Set(name, item)
			}
		}
	}
	if _, err := h.service.VerifyNotification(c.Request.Context(), provider, values); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

func (h *Handler) ConfigureGateway(c *gin.Context) {
	var input GatewayInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "provider, name, baseUrl and merchantId are required"})
		return
	}
	if h.service == nil {
		return
	}
	item, err := h.service.Store().SaveGateway(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "PAYMENT_GATEWAY_SAVE_FAILED", "message": err.Error()})
		return
	}
	item.EncryptedKey = ""
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "saved", "data": item})
}

func (h *Handler) Gateway(c *gin.Context) {
	if h.service == nil {
		return
	}
	provider := c.DefaultQuery("provider", ProviderEpay)
	item, err := h.service.Store().GetGateway(c.Request.Context(), provider)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "PAYMENT_GATEWAY_NOT_FOUND", "message": "payment gateway is not configured"})
		return
	}
	item.EncryptedKey = ""
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
}

func (h *Handler) CloseExpired(c *gin.Context) {
	if h.service == nil {
		return
	}
	count, err := h.service.Store().CloseExpired(c.Request.Context(), nowUTC())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "failed to close orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "closed", "data": gin.H{"count": count}})
}

func (h *Handler) Refund(c *gin.Context) {
	orderNo := c.Param("orderNo")
	var input RefundInput
	_ = c.ShouldBindJSON(&input)
	amount := input.AmountCents
	if value := c.Query("amountCents"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "amountCents is invalid"})
			return
		}
		amount = parsed
	}
	if amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "amountCents is required"})
		return
	}
	if h.service == nil {
		return
	}
	refundNo := input.RefundNo
	if refundNo == "" {
		refundNo = orderNo + "-R" + strconv.FormatInt(nowUTC().UnixNano(), 10)
	}
	reason := input.Reason
	if reason == "" {
		reason = c.Query("reason")
	}
	item, err := h.service.Store().RecordRefund(c.Request.Context(), orderNo, amount, reason, refundNo)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrOrderNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"code": "REFUND_FAILED", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "refunded", "data": item})
}

func (h *Handler) Refunds(c *gin.Context) {
	if h.service == nil {
		return
	}
	items, err := h.service.Store().ListRefunds(c.Request.Context(), c.Param("orderNo"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrOrderNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"code": "REFUNDS_LOAD_FAILED", "message": "failed to load refunds"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items})
}

func (h *Handler) Reconciliation(c *gin.Context) {
	if h.service == nil {
		return
	}
	items, err := h.service.Store().Reconcile(c.Request.Context(), nowUTC())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "RECONCILIATION_FAILED", "message": "failed to reconcile orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"issues": items, "count": len(items)}})
}

func currentUser(c *gin.Context) (auth.User, bool) {
	value, exists := c.Get("currentUser")
	user, ok := value.(auth.User)
	if !exists || !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "authentication required"})
		return auth.User{}, false
	}
	return user, true
}

func nowUTC() time.Time { return time.Now().UTC() }
