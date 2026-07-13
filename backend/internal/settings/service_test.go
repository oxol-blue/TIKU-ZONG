package settings

import "testing"

func TestValidateUpdateInput(t *testing.T) {
	valid, err := validateUpdateInput(UpdateInput{SiteName: "  题库  ", SupportURL: "https://example.com/help", MaintenanceNotice: " notice "})
	if err != nil {
		t.Fatalf("expected valid settings, got %v", err)
	}
	if valid.SiteName != "题库" || valid.MaintenanceNotice != "notice" {
		t.Fatalf("expected trimmed values, got %#v", valid)
	}
	invalid := []UpdateInput{
		{SiteName: ""},
		{SiteName: "题库", SupportURL: "javascript:alert(1)"},
	}
	for _, input := range invalid {
		if _, err := validateUpdateInput(input); err != ErrInvalidInput {
			t.Fatalf("expected invalid settings for %#v, got %v", input, err)
		}
	}
}
