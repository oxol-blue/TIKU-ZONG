package settings

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidInput = errors.New("invalid system settings")

type Service struct{ store *Store }

func NewService(store *Store) *Service { return &Service{store: store} }

func (s *Service) Public(ctx context.Context) (PublicConfig, error) {
	values, err := s.store.GetMany(ctx, []string{"site_name", "support_url", "maintenance_notice", "registration_enabled"})
	if err != nil {
		return PublicConfig{}, err
	}
	return PublicConfig{SiteName: fallback(values["site_name"], "题库调用系统"), SupportURL: values["support_url"], MaintenanceNotice: values["maintenance_notice"], RegistrationEnabled: values["registration_enabled"] != "false"}, nil
}

func (s *Service) Admin(ctx context.Context) (AdminConfig, error) {
	config, err := s.Public(ctx)
	return AdminConfig{PublicConfig: config}, err
}

func (s *Service) RegistrationEnabled(ctx context.Context) (bool, error) {
	value, err := s.store.Get(ctx, "registration_enabled")
	return value != "false", err
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (AdminConfig, error) {
	input, err := validateUpdateInput(input)
	if err != nil {
		return AdminConfig{}, err
	}
	if err := s.store.Put(ctx, map[string]string{
		"site_name": input.SiteName, "support_url": input.SupportURL, "maintenance_notice": input.MaintenanceNotice,
		"registration_enabled": map[bool]string{true: "true", false: "false"}[input.RegistrationEnabled],
	}); err != nil {
		return AdminConfig{}, err
	}
	return s.Admin(ctx)
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.SiteName = strings.TrimSpace(input.SiteName)
	input.SupportURL = strings.TrimSpace(input.SupportURL)
	input.MaintenanceNotice = strings.TrimSpace(input.MaintenanceNotice)
	if input.SiteName == "" || len(input.SiteName) > 64 || len(input.MaintenanceNotice) > 500 {
		return UpdateInput{}, ErrInvalidInput
	}
	if input.SupportURL != "" {
		parsed, err := url.ParseRequestURI(input.SupportURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return UpdateInput{}, ErrInvalidInput
		}
	}
	return input, nil
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
