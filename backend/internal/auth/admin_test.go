package auth

import (
	"context"
	"testing"
)

func TestAdminUserUpdateValidation(t *testing.T) {
	service := &Service{}
	if err := service.UpdateUserStatus(context.Background(), 7, 7, 0); err != ErrSelfModification {
		t.Fatalf("expected self-disable protection, got %v", err)
	}
	if err := service.UpdateUserRole(context.Background(), 7, 7, RoleUser); err != ErrSelfModification {
		t.Fatalf("expected self-demotion protection, got %v", err)
	}
	if err := service.UpdateUserRole(context.Background(), 7, 8, "operator"); err != ErrInvalidInput {
		t.Fatalf("expected invalid role error, got %v", err)
	}
	if err := service.UpdateUserStatus(context.Background(), 7, 8, 2); err != ErrInvalidInput {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestNormalizePage(t *testing.T) {
	if page, size := normalizePage(0, 0); page != 1 || size != 20 {
		t.Fatalf("unexpected defaults: page=%d size=%d", page, size)
	}
	if page, size := normalizePage(2, 500); page != 2 || size != 100 {
		t.Fatalf("unexpected limits: page=%d size=%d", page, size)
	}
}

func TestPasswordChangeValidation(t *testing.T) {
	invalid := [][2]string{{"", "new-password"}, {"old", "short"}, {"old", ""}, {"same-password", "same-password"}}
	for _, pair := range invalid {
		if err := validatePasswordChange(pair[0], pair[1]); err != ErrInvalidInput {
			t.Fatalf("expected invalid password input for %q, got %v", pair, err)
		}
	}
	if err := validatePasswordChange("old-password", "new-password"); err != nil {
		t.Fatalf("expected valid password input, got %v", err)
	}
}
