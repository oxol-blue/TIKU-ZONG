package payment

import (
	"net/url"
	"strings"
	"testing"
)

func TestEpayPaymentURLAndSignature(t *testing.T) {
	adapter := EpayAdapter{Decrypt: func(string) (string, error) { return "secret", nil }}
	gateway := Gateway{BaseURL: "https://pay.example.test", MerchantID: "10001", EncryptedKey: "ciphertext"}
	order := Order{OrderNo: "202607130001", PackageName: "次数套餐", PayableCents: 1234}
	paymentURL, err := adapter.BuildPaymentURL(gateway, order, PaymentRequest{OrderNo: order.OrderNo, Notify: "https://app.test/notify", Return: "https://app.test/result"})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(paymentURL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(parsed.Path, "/submit.php") {
		t.Fatalf("expected submit.php endpoint, got %s", parsed.Path)
	}
	if got := parsed.Query().Get("money"); got != "12.34" {
		t.Fatalf("expected money 12.34, got %q", got)
	}
	values := make(map[string]string)
	for name := range parsed.Query() {
		values[name] = parsed.Query().Get(name)
	}
	notification, err := adapter.VerifyNotification(gateway, values)
	if err != nil {
		t.Fatal(err)
	}
	if notification.OrderNo != order.OrderNo || notification.AmountCents != order.PayableCents {
		t.Fatalf("unexpected notification: %+v", notification)
	}
	values["sign"] = "bad"
	if _, err := adapter.VerifyNotification(gateway, values); err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestParseMoneyCents(t *testing.T) {
	cases := map[string]int{"1": 100, "1.2": 120, "1.23": 123, "0.00": 0}
	for input, expected := range cases {
		got, err := parseMoneyCents(input)
		if err != nil || got != expected {
			t.Fatalf("parseMoneyCents(%q) = %d, %v; want %d", input, got, err, expected)
		}
	}
}
