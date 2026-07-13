package auth

import (
	"context"
	"testing"
)

func TestCreateInviteValidation(t *testing.T) {
	service := &Service{store: &Store{}}
	if _, err := service.CreateInvite(context.Background(), 1, CreateInviteInput{Code: "", MaxUses: 1}); err != ErrInvalidInput {
		t.Fatalf("expected empty code validation error, got %v", err)
	}
	if _, err := service.CreateInvite(context.Background(), 1, CreateInviteInput{Code: "ABC", MaxUses: -1}); err != ErrInvalidInput {
		t.Fatalf("expected negative usage validation error, got %v", err)
	}
}
