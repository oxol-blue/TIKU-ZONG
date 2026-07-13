package ai

import (
	"context"
	"testing"
)

func TestAdminAnswerStatusValidation(t *testing.T) {
	service := &Service{}
	if err := service.UpdateAnswerStatus(context.Background(), 1, 2); err == nil {
		t.Fatal("expected invalid AI answer status")
	}
}

func TestModelStatusValidation(t *testing.T) {
	service := &Service{}
	if err := service.UpdateModelStatus(context.Background(), 1, 2); err == nil {
		t.Fatal("expected invalid AI model status")
	}
}

func TestModelInputValidation(t *testing.T) {
	if err := validateModelInput(CreateModelInput{ProviderID: 1, Name: "demo", BillingMode: BillingModeToken}); err == nil {
		t.Fatal("expected missing token unit to be rejected")
	}
}

func TestAnswerPageLimits(t *testing.T) {
	if page, size := normalizePage(0, 0); page != 1 || size != 20 {
		t.Fatalf("unexpected defaults: page=%d size=%d", page, size)
	}
	if page, size := normalizePage(4, 1000); page != 4 || size != 100 {
		t.Fatalf("unexpected limits: page=%d size=%d", page, size)
	}
}
