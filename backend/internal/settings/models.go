package settings

import "time"

type Item struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	IsPublic  bool      `json:"isPublic"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PublicConfig struct {
	SiteName            string `json:"siteName"`
	SupportURL          string `json:"supportUrl"`
	MaintenanceNotice   string `json:"maintenanceNotice"`
	RegistrationEnabled bool   `json:"registrationEnabled"`
}

type UpdateInput struct {
	SiteName            string `json:"siteName"`
	SupportURL          string `json:"supportUrl"`
	MaintenanceNotice   string `json:"maintenanceNotice"`
	RegistrationEnabled bool   `json:"registrationEnabled"`
}

type AdminConfig struct {
	PublicConfig
}
