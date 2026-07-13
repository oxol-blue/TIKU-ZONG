package ai

import "time"

type CreateProviderInput struct {
	Name    string `json:"name" binding:"required"`
	BaseURL string `json:"baseUrl" binding:"required"`
	APIKey  string `json:"apiKey" binding:"required"`
}

type CreateModelInput struct {
	ProviderID     uint64 `json:"providerId" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Priority       int    `json:"priority"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	AIChargeCount  int    `json:"aiChargeCount"`
}

type Model struct {
	ID             uint64 `json:"id"`
	ProviderID     uint64 `json:"providerId"`
	ProviderName   string `json:"providerName"`
	BaseURL        string `json:"-"`
	EncryptedKey   string `json:"-"`
	KeyConfigured  bool   `json:"keyConfigured"`
	Name           string `json:"name"`
	Priority       int    `json:"priority"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	AIChargeCount  int    `json:"aiChargeCount"`
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
