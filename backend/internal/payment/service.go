package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	store   *Store
	public  string
	adapter GatewayAdapter
}

func NewService(store *Store, publicBaseURL string) *Service {
	return &Service{store: store, public: strings.TrimRight(publicBaseURL, "/"), adapter: EpayAdapter{Decrypt: func(value string) (string, error) { return store.DecryptKey(value) }}}
}

func (s *Service) CreateOrder(ctx context.Context, userID uint64, input CreateOrderInput) (Order, string, error) {
	provider := input.Provider
	if provider == "" {
		provider = ProviderEpay
	}
	orderNo, err := newOrderNo()
	if err != nil {
		return Order{}, "", err
	}
	order, err := s.store.CreateOrder(ctx, userID, input.PackageID, provider, input.CouponCode, orderNo, time.Now().UTC().Add(30*time.Minute))
	if err != nil {
		return Order{}, "", err
	}
	if order.PayableCents == 0 {
		paid, err := s.store.MarkPaidAndGrant(ctx, Notification{OrderNo: order.OrderNo, Status: "SUCCESS", AmountCents: 0})
		if err != nil {
			return Order{}, "", err
		}
		return paid, "", nil
	}
	gateway, err := s.store.GetGateway(ctx, provider)
	if err != nil {
		return order, "", errors.New("payment gateway is not configured")
	}
	if gateway.Enabled != 1 {
		return order, "", errors.New("payment gateway is disabled")
	}
	paymentURL, err := s.adapter.BuildPaymentURL(gateway, order, PaymentRequest{OrderNo: order.OrderNo, Notify: s.public + "/api/payment/notify/" + url.PathEscape(provider), Return: s.public + "/payment/result?order_no=" + url.QueryEscape(order.OrderNo)})
	if err != nil {
		return order, "", err
	}
	return order, paymentURL, nil
}

func (s *Service) VerifyNotification(ctx context.Context, provider string, values url.Values) (Order, error) {
	gateway, err := s.store.GetGateway(ctx, provider)
	if err != nil {
		return Order{}, err
	}
	flat := make(map[string]string, len(values))
	for name := range values {
		flat[name] = values.Get(name)
	}
	notification, err := s.adapter.VerifyNotification(gateway, flat)
	if err != nil {
		return Order{}, err
	}
	if strings.ToUpper(notification.Status) != "TRADE_SUCCESS" && strings.ToUpper(notification.Status) != "SUCCESS" {
		return Order{}, errors.New("payment is not successful")
	}
	return s.store.MarkPaidAndGrant(ctx, notification)
}

func (s *Service) Store() *Store { return s.store }

func newOrderNo() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return time.Now().UTC().Format("20060102150405") + hex.EncodeToString(bytes), nil
}

func Money(cents int) string { return strconv.FormatFloat(float64(cents)/100, 'f', 2, 64) }
