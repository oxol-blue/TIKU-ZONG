package captcha

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateAndVerifyCaptcha(t *testing.T) {
	store := NewStore()
	id, image, err := store.Generate()
	if err != nil || id == "" || !strings.HasPrefix(image, "data:image/svg+xml;base64,") {
		t.Fatalf("unexpected captcha response: %q, %v", id, err)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(image, "data:image/svg+xml;base64,"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	start := strings.Index(text, `letter-spacing="4" fill="#2563eb">`)
	if start >= 0 {
		start += len(`letter-spacing="4" fill="#2563eb">`)
	}
	end := strings.Index(text[start:], "</text>")
	if start < 0 || end < 0 {
		t.Fatal("captcha SVG text is missing")
	}
	code := text[start : start+end]
	if !store.Verify(id, code) || store.Verify(id, code) {
		t.Fatal("captcha should verify once")
	}
}
