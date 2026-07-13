package secret

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	value, err := Encrypt("provider-secret", "master-secret")
	if err != nil {
		t.Fatal(err)
	}
	if value == "provider-secret" || value == "" {
		t.Fatalf("expected encrypted value, got %q", value)
	}
	decoded, err := Decrypt(value, "master-secret")
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "provider-secret" {
		t.Fatalf("expected original secret, got %q", decoded)
	}
}
