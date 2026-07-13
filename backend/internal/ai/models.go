package ai

import "time"

const (
	BillingModeFixed = "fixed"
	BillingModeToken = "token"
	BillingModeCost  = "cost"
)

type CreateProviderInput struct {
	Name    string `json:"name" binding:"required"`
	BaseURL string `json:"baseUrl" binding:"required"`
	APIKey  string `json:"apiKey" binding:"required"`
}

type CreateModelInput struct {
	ProviderID                uint64 `json:"providerId" binding:"required"`
	Name                      string `json:"name" binding:"required"`
	Priority                  int    `json:"priority"`
	TimeoutSeconds            int    `json:"timeoutSeconds"`
	AIChargeCount             int    `json:"aiChargeCount"`
	BillingMode               string `json:"billingMode"`
	TokenUnit                 int    `json:"tokenUnit"`
	CostPerMillionTokensCents int    `json:"costPerMillionTokensCents"`
	CostMarkupPercent         int    `json:"costMarkupPercent"`
	CostUnitCents             int    `json:"costUnitCents"`
}

type Model struct {
	ID                        uint64 `json:"id"`
	ProviderID                uint64 `json:"providerId"`
	ProviderName              string `json:"providerName"`
	BaseURL                   string `json:"-"`
	EncryptedKey              string `json:"-"`
	KeyConfigured             bool   `json:"keyConfigured"`
	Name                      string `json:"name"`
	Priority                  int    `json:"priority"`
	TimeoutSeconds            int    `json:"timeoutSeconds"`
	AIChargeCount             int    `json:"aiChargeCount"`
	BillingMode               string `json:"billingMode"`
	TokenUnit                 int    `json:"tokenUnit"`
	CostPerMillionTokensCents int    `json:"costPerMillionTokensCents"`
	CostMarkupPercent         int    `json:"costMarkupPercent"`
	CostUnitCents             int    `json:"costUnitCents"`
}

type Answer struct {
	QuestionHash string
	Question     string
	Type         string
	Text         string
	RawResponse  string
	Provider     string
	Model        string
	TokenCount   int
	ChargeCount  int
	Elapsed      time.Duration
}

type AdminAnswer struct {
	ID           uint64    `json:"id"`
	QuestionHash string    `json:"questionHash"`
	Question     string    `json:"question"`
	Type         string    `json:"type"`
	Text         string    `json:"answer"`
	Prompt       string    `json:"prompt"`
	RawResponse  string    `json:"rawResponse"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	TokenCount   int       `json:"tokenCount"`
	Elapsed      int64     `json:"elapsedMicros"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AnswerPage struct {
	Items    []AdminAnswer `json:"items"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Total    int           `json:"total"`
}
