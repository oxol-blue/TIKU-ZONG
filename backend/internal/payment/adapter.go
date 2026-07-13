package payment

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type EpayAdapter struct {
	Decrypt func(string) (string, error)
}

func (a EpayAdapter) BuildPaymentURL(gateway Gateway, order Order, request PaymentRequest) (string, error) {
	key, err := a.Decrypt(gateway.EncryptedKey)
	if err != nil || key == "" {
		return "", errors.New("payment gateway secret is unavailable")
	}
	values := map[string]string{
		"pid":          gateway.MerchantID,
		"type":         "alipay",
		"out_trade_no": order.OrderNo,
		"notify_url":   request.Notify,
		"return_url":   request.Return,
		"name":         order.PackageName,
		"money":        fmt.Sprintf("%.2f", float64(order.PayableCents)/100),
		"sign_type":    "MD5",
	}
	values["sign"] = sign(values, key)
	endpoint := strings.TrimRight(gateway.BaseURL, "/")
	if !strings.HasSuffix(endpoint, ".php") {
		endpoint += "/submit.php"
	}
	query := url.Values{}
	for name, value := range values {
		query.Set(name, value)
	}
	return endpoint + "?" + query.Encode(), nil
}

func (a EpayAdapter) VerifyNotification(gateway Gateway, values map[string]string) (Notification, error) {
	key, err := a.Decrypt(gateway.EncryptedKey)
	if err != nil || key == "" {
		return Notification{}, errors.New("payment gateway secret is unavailable")
	}
	provided := strings.ToLower(strings.TrimSpace(values["sign"]))
	expected := sign(values, key)
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return Notification{}, errors.New("invalid payment signature")
	}
	status := strings.ToUpper(strings.TrimSpace(values["trade_status"]))
	if status == "" {
		status = strings.ToUpper(strings.TrimSpace(values["status"]))
	}
	amount, err := parseMoneyCents(values["money"])
	if err != nil {
		return Notification{}, err
	}
	return Notification{OrderNo: values["out_trade_no"], ProviderTradeNo: values["trade_no"], Status: status, AmountCents: amount}, nil
}

func sign(values map[string]string, key string) string {
	keys := make([]string, 0, len(values))
	for name, value := range values {
		if name == "sign" || name == "sign_type" || value == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+1)
	for _, name := range keys {
		parts = append(parts, name+"="+values[name])
	}
	parts = append(parts, "key="+key)
	digest := md5.Sum([]byte(strings.Join(parts, "&")))
	return hex.EncodeToString(digest[:])
}

func parseMoneyCents(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("payment amount is missing")
	}
	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.Atoi(parts[0])
	if err != nil || whole < 0 {
		return 0, errors.New("invalid payment amount")
	}
	decimal := "00"
	if len(parts) == 2 {
		decimal = parts[1]
	}
	if len(decimal) > 2 || strings.Trim(decimal, "0123456789") != "" {
		return 0, errors.New("invalid payment amount")
	}
	decimal = (decimal + "00")[:2]
	return whole*100 + atoiDigits(decimal), nil
}

func atoiDigits(value string) int {
	result := 0
	for _, char := range value {
		result = result*10 + int(char-'0')
	}
	return result
}
