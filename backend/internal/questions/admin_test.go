package questions

import (
	"context"
	"testing"
)

func TestAdminQuestionStatusValidation(t *testing.T) {
	service := &Service{}
	if err := service.UpdateStatus(context.Background(), 1, 2); err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestQuestionAdminPageLimits(t *testing.T) {
	if page, size := normalizePage(0, 0); page != 1 || size != 20 {
		t.Fatalf("unexpected defaults: page=%d size=%d", page, size)
	}
	if page, size := normalizePage(3, 1000); page != 3 || size != 100 {
		t.Fatalf("unexpected limits: page=%d size=%d", page, size)
	}
}
