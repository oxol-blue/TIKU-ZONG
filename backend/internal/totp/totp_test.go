package totp

import (
	"testing"
	"time"
)

func TestRFC6238StyleVerification(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	now := time.Unix(59, 0).UTC()
	code := Generate(secret, now)
	if code == "" || !Verify(secret, code, now) || Verify(secret, "000000", now) {
		t.Fatalf("unexpected TOTP result: %q", code)
	}
}
