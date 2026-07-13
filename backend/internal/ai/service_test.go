package ai

import (
	"context"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	prompt := buildPrompt("题目", "single", []string{"A. 北京", "B. 上海"})
	if prompt == "" || prompt == "题目" {
		t.Fatalf("expected prompt context, got %q", prompt)
	}
}

func TestNewServiceWithQueueNormalizesInvalidSettings(t *testing.T) {
	service := NewServiceWithQueue(nil, 0, 0)
	if cap(service.jobs) != 1 {
		t.Fatalf("expected minimum queue size 1, got %d", cap(service.jobs))
	}
	if service.workers != 1 {
		t.Fatalf("expected minimum worker count 1, got %d", service.workers)
	}
}

func TestSolveReturnsQueueFullWithoutBlocking(t *testing.T) {
	service := &Service{jobs: make(chan solveJob, 1)}
	service.jobs <- solveJob{}
	_, err := service.Solve(context.Background(), "question", "", nil)
	if err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestChargeCountModes(t *testing.T) {
	tests := []struct {
		name   string
		model  Model
		tokens int
		want   int
	}{
		{name: "fixed", model: Model{BillingMode: BillingModeFixed, AIChargeCount: 3}, tokens: 9999, want: 3},
		{name: "token", model: Model{BillingMode: BillingModeToken, AIChargeCount: 2, TokenUnit: 1000}, tokens: 2500, want: 6},
		{name: "cost", model: Model{BillingMode: BillingModeCost, AIChargeCount: 2, CostPerMillionTokensCents: 1000, CostMarkupPercent: 50, CostUnitCents: 1}, tokens: 1000, want: 2},
		{name: "missing usage falls back", model: Model{BillingMode: BillingModeToken, AIChargeCount: 4, TokenUnit: 1000}, tokens: 0, want: 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := chargeCount(test.model, test.tokens); got != test.want {
				t.Fatalf("chargeCount() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestValidateModelInput(t *testing.T) {
	if err := validateModelInput(CreateModelInput{ProviderID: 1, Name: "model", BillingMode: BillingModeToken}); err == nil {
		t.Fatal("expected token mode without token unit to fail")
	}
	if err := validateModelInput(CreateModelInput{ProviderID: 1, Name: "model", BillingMode: BillingModeCost, CostPerMillionTokensCents: 100, CostUnitCents: 1}); err != nil {
		t.Fatalf("expected valid cost mode, got %v", err)
	}
}
