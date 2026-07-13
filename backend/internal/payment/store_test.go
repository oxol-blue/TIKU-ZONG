package payment

import (
	"testing"
	"time"
)

func TestPackageInstanceEntitlementRules(t *testing.T) {
	duration := int64(3600)
	if got := remainingForPackage(0, &duration); got != -1 {
		t.Fatalf("time package should have unlimited remaining count, got %d", got)
	}
	if got := remainingForPackage(12, &duration); got != 12 {
		t.Fatalf("time-count package should preserve its count, got %d", got)
	}
	if got := remainingForPackage(5, nil); got != 5 {
		t.Fatalf("count package should preserve its count, got %d", got)
	}
	start := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	expires := expiresForPackage(start, &duration)
	if expires == nil || !expires.Equal(start.Add(time.Hour)) {
		t.Fatalf("unexpected expiration: %#v", expires)
	}
	if expiresForPackage(start, nil) != nil {
		t.Fatal("count package should not have an expiration")
	}
}
