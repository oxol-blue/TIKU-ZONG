package payment

import "time"

const (
	ProviderEpay = "epay"

	OrderPending         = "pending"
	OrderPaid            = "paid"
	OrderClosed          = "closed"
	OrderRefunded        = "refunded"
	OrderPartialRefunded = "partial_refunded"
)

type GatewayInput struct {
	Provider   string `json:"provider" binding:"required"`
	Name       string `json:"name" binding:"required"`
	BaseURL    string `json:"baseUrl" binding:"required"`
	MerchantID string `json:"merchantId" binding:"required"`
	SecretKey  string `json:"secretKey" binding:"required"`
	Enabled    bool   `json:"enabled"`
}

type Gateway struct {
	ID            uint64 `json:"id"`
	Provider      string `json:"provider"`
	Name          string `json:"name"`
	BaseURL       string `json:"baseUrl"`
	MerchantID    string `json:"merchantId"`
	EncryptedKey  string `json:"-"`
	KeyConfigured bool   `json:"keyConfigured"`
	Enabled       int    `json:"enabled"`
}

type Order struct {
	ID                uint64     `json:"id"`
	OrderNo           string     `json:"orderNo"`
	UserID            uint64     `json:"userId"`
	PackageID         uint64     `json:"packageId"`
	PackageName       string     `json:"packageName"`
	Provider          string     `json:"provider"`
	AmountCents       int        `json:"amountCents"`
	PayableCents      int        `json:"payableCents"`
	RefundedCents     int        `json:"refundedCents"`
	Status            string     `json:"status"`
	ProviderTradeNo   string     `json:"providerTradeNo,omitempty"`
	PackageInstanceID uint64     `json:"packageInstanceId,omitempty"`
	ExpiresAt         time.Time  `json:"expiresAt"`
	PaidAt            *time.Time `json:"paidAt,omitempty"`
	ClosedAt          *time.Time `json:"closedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type CreateOrderInput struct {
	PackageID uint64 `json:"packageId" binding:"required"`
	Provider  string `json:"provider"`
}

type RefundInput struct {
	AmountCents int    `json:"amountCents" binding:"required"`
	Reason      string `json:"reason"`
}

type PaymentRequest struct {
	OrderNo string
	Notify  string
	Return  string
}

type Notification struct {
	OrderNo         string
	ProviderTradeNo string
	Status          string
	AmountCents     int
}

type GatewayAdapter interface {
	BuildPaymentURL(gateway Gateway, order Order, request PaymentRequest) (string, error)
	VerifyNotification(gateway Gateway, values map[string]string) (Notification, error)
}
