package billing

import (
	"strings"
	"testing"
)

func TestPackageValidation(t *testing.T) {
	duration := int64(3600)
	tests := []struct {
		name  string
		input CreatePackageInput
	}{
		{name: "unknown type", input: CreatePackageInput{Type: "unknown", TotalCount: 1}},
		{name: "missing duration", input: CreatePackageInput{Type: PackageTime, TotalCount: 0}},
		{name: "missing count", input: CreatePackageInput{Type: PackageCount}},
		{name: "negative price", input: CreatePackageInput{Type: PackageCount, TotalCount: 1, PriceCents: -1}},
		{name: "invalid flags", input: CreatePackageInput{Type: PackageCount, TotalCount: 1, IsFree: 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validatePackage(test.input.Type, test.input.DurationSeconds, test.input.TotalCount, test.input.AICount, test.input.PriceCents, test.input.LimitCount, test.input.IsTrial, test.input.IsFree); err == nil {
				t.Fatal("expected package validation error")
			}
		})
	}
	if err := validatePackage(PackageTime, &duration, 0, 0, 0, 0, 0, 0); err != nil {
		t.Fatalf("expected valid time package, got %v", err)
	}
}

func TestCouponValidation(t *testing.T) {
	tests := []CreateCouponInput{
		{Code: "", DiscountType: "percent", DiscountValue: 10},
		{Code: "SAVE", DiscountType: "unknown", DiscountValue: 10},
		{Code: "SAVE", DiscountType: "percent", DiscountValue: 101},
		{Code: "SAVE", DiscountType: "fixed", DiscountValue: 0},
		{Code: "SAVE", DiscountType: "fixed", DiscountValue: 10, TotalLimit: -1},
	}
	for _, input := range tests {
		input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
		if err := validateCoupon(input); err == nil {
			t.Fatalf("expected coupon validation error for %#v", input)
		}
	}
}

func TestStatusValidation(t *testing.T) {
	if err := validateStatus(2); err == nil {
		t.Fatal("expected package status validation error")
	}
	if err := validateStatus(-1); err == nil {
		t.Fatal("expected coupon status validation error")
	}
}
