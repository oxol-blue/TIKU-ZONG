package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oxol-blue/TIKU-ZONG/backend/internal/secret"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderClosed   = errors.New("order is closed")
	ErrOrderPaid     = errors.New("order is already paid")
)

type Store struct {
	db     *sql.DB
	secret string
}

func NewStore(db *sql.DB, masterSecret string) *Store { return &Store{db: db, secret: masterSecret} }

func (s *Store) DecryptKey(value string) (string, error) { return secret.Decrypt(value, s.secret) }

func (s *Store) SaveGateway(ctx context.Context, input GatewayInput) (Gateway, error) {
	if input.Provider != ProviderEpay {
		return Gateway{}, errors.New("unsupported payment provider")
	}
	ciphertext, err := secret.Encrypt(input.SecretKey, s.secret)
	if err != nil {
		return Gateway{}, err
	}
	enabled := 0
	if input.Enabled {
		enabled = 1
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO payment_gateways (provider, name, base_url, merchant_id, secret_ciphertext, enabled) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name = VALUES(name), base_url = VALUES(base_url), merchant_id = VALUES(merchant_id), secret_ciphertext = VALUES(secret_ciphertext), enabled = VALUES(enabled), updated_at = CURRENT_TIMESTAMP(6)`, input.Provider, input.Name, input.BaseURL, input.MerchantID, ciphertext, enabled)
	if err != nil {
		return Gateway{}, fmt.Errorf("save payment gateway: %w", err)
	}
	return s.GetGateway(ctx, input.Provider)
}

func (s *Store) GetGateway(ctx context.Context, provider string) (Gateway, error) {
	var item Gateway
	err := s.db.QueryRowContext(ctx, `SELECT id, provider, name, base_url, merchant_id, secret_ciphertext, enabled FROM payment_gateways WHERE provider = ?`, provider).
		Scan(&item.ID, &item.Provider, &item.Name, &item.BaseURL, &item.MerchantID, &item.EncryptedKey, &item.Enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return Gateway{}, ErrOrderNotFound
	}
	item.KeyConfigured = item.EncryptedKey != ""
	return item, err
}

func (s *Store) CreateOrder(ctx context.Context, userID, packageID uint64, provider, couponCode, orderNo string, expiresAt time.Time) (Order, error) {
	if provider == "" {
		provider = ProviderEpay
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()
	var packageName string
	var price, status, limitCount int
	var isFree, isTrial int
	err = tx.QueryRowContext(ctx, `SELECT name, price_cents, status, limit_count, is_free, is_trial FROM packages WHERE id = ? FOR UPDATE`, packageID).Scan(&packageName, &price, &status, &limitCount, &isFree, &isTrial)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, errors.New("package not found")
	}
	if err != nil {
		return Order{}, err
	}
	if status != 1 {
		return Order{}, errors.New("package is not available")
	}
	if price < 0 {
		return Order{}, errors.New("package price is invalid")
	}
	if isFree == 1 {
		price = 0
	}
	if isTrial == 1 && limitCount == 0 {
		limitCount = 1
	}
	if limitCount > 0 {
		var purchased int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM package_instances WHERE user_id = ? AND package_id = ?`, userID, packageID).Scan(&purchased); err != nil {
			return Order{}, err
		}
		if purchased >= limitCount {
			return Order{}, errors.New("package purchase limit reached")
		}
	}
	couponID, discountCents := uint64(0), 0
	if couponCode != "" && price > 0 {
		couponCode = strings.ToUpper(strings.TrimSpace(couponCode))
		var couponStatus, totalLimit, usedCount, reservedCount, discountValue int
		var discountType string
		var couponExpires *time.Time
		err = tx.QueryRowContext(ctx, `SELECT id, discount_type, discount_value, total_limit, used_count, reserved_count, expires_at, status FROM coupons WHERE code = ? FOR UPDATE`, couponCode).
			Scan(&couponID, &discountType, &discountValue, &totalLimit, &usedCount, &reservedCount, &couponExpires, &couponStatus)
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, errors.New("coupon not found")
		}
		if err != nil {
			return Order{}, err
		}
		if couponStatus != 1 || (couponExpires != nil && !time.Now().UTC().Before(*couponExpires)) || (totalLimit > 0 && usedCount+reservedCount >= totalLimit) {
			return Order{}, errors.New("coupon is unavailable")
		}
		if discountType == "fixed" {
			discountCents = discountValue
		} else if discountType == "percent" {
			discountCents = price * discountValue / 100
		}
		if discountCents > price {
			discountCents = price
		}
		if _, err = tx.ExecContext(ctx, `UPDATE coupons SET reserved_count = reserved_count + 1 WHERE id = ?`, couponID); err != nil {
			return Order{}, err
		}
	}
	payable := price - discountCents
	result, err := tx.ExecContext(ctx, `INSERT INTO payment_orders (order_no, user_id, package_id, provider, coupon_id, coupon_code, amount_cents, payable_cents, discount_cents, status, expires_at) VALUES (?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?)`, orderNo, userID, packageID, provider, couponID, couponCode, price, payable, discountCents, OrderPending, expiresAt)
	if err != nil {
		return Order{}, fmt.Errorf("create order: %w", err)
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		return Order{}, err
	}
	if couponID != 0 {
		if _, err = tx.ExecContext(ctx, `INSERT INTO coupon_reservations (coupon_id, order_id) VALUES (?, ?)`, couponID, orderID); err != nil {
			return Order{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}
	_ = packageName
	return s.GetOrder(ctx, orderNo)
}

func (s *Store) GetOrder(ctx context.Context, orderNo string) (Order, error) {
	var item Order
	err := s.db.QueryRowContext(ctx, `SELECT o.id, o.order_no, o.user_id, o.package_id, p.name, o.provider, COALESCE(o.coupon_id, 0), o.coupon_code, o.amount_cents, o.payable_cents, o.discount_cents, o.refunded_cents, o.status, o.provider_trade_no, o.package_instance_id, o.expires_at, o.paid_at, o.closed_at, o.created_at FROM payment_orders o JOIN packages p ON p.id = o.package_id WHERE o.order_no = ?`, orderNo).
		Scan(&item.ID, &item.OrderNo, &item.UserID, &item.PackageID, &item.PackageName, &item.Provider, &item.CouponID, &item.CouponCode, &item.AmountCents, &item.PayableCents, &item.DiscountCents, &item.RefundedCents, &item.Status, &item.ProviderTradeNo, &item.PackageInstanceID, &item.ExpiresAt, &item.PaidAt, &item.ClosedAt, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	return item, err
}

func (s *Store) ListOrders(ctx context.Context, userID uint64) ([]Order, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT o.id, o.order_no, o.user_id, o.package_id, p.name, o.provider, COALESCE(o.coupon_id, 0), o.coupon_code, o.amount_cents, o.payable_cents, o.discount_cents, o.refunded_cents, o.status, o.provider_trade_no, o.package_instance_id, o.expires_at, o.paid_at, o.closed_at, o.created_at FROM payment_orders o JOIN packages p ON p.id = o.package_id WHERE o.user_id = ? ORDER BY o.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Order, 0)
	for rows.Next() {
		var item Order
		if err := rows.Scan(&item.ID, &item.OrderNo, &item.UserID, &item.PackageID, &item.PackageName, &item.Provider, &item.CouponID, &item.CouponCode, &item.AmountCents, &item.PayableCents, &item.DiscountCents, &item.RefundedCents, &item.Status, &item.ProviderTradeNo, &item.PackageInstanceID, &item.ExpiresAt, &item.PaidAt, &item.ClosedAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) MarkPaidAndGrant(ctx context.Context, notification Notification) (Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()
	var item Order
	var packageStatus int
	var duration *int64
	var totalCount, aiCount int
	err = tx.QueryRowContext(ctx, `SELECT o.id, o.order_no, o.user_id, o.package_id, p.name, o.provider, COALESCE(o.coupon_id, 0), o.coupon_code, o.amount_cents, o.payable_cents, o.discount_cents, o.refunded_cents, o.status, o.provider_trade_no, o.package_instance_id, o.expires_at, o.paid_at, o.closed_at, o.created_at, p.status, p.duration_seconds, p.total_count, p.ai_count FROM payment_orders o JOIN packages p ON p.id = o.package_id WHERE o.order_no = ? FOR UPDATE`, notification.OrderNo).
		Scan(&item.ID, &item.OrderNo, &item.UserID, &item.PackageID, &item.PackageName, &item.Provider, &item.CouponID, &item.CouponCode, &item.AmountCents, &item.PayableCents, &item.DiscountCents, &item.RefundedCents, &item.Status, &item.ProviderTradeNo, &item.PackageInstanceID, &item.ExpiresAt, &item.PaidAt, &item.ClosedAt, &item.CreatedAt, &packageStatus, &duration, &totalCount, &aiCount)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, err
	}
	if item.Status == OrderPaid || item.Status == OrderPartialRefunded || item.Status == OrderRefunded {
		return item, nil
	}
	if item.Status != OrderPending {
		return Order{}, ErrOrderClosed
	}
	if notification.AmountCents != item.PayableCents {
		return Order{}, errors.New("payment amount mismatch")
	}
	if packageStatus != 1 {
		return Order{}, errors.New("package is no longer available")
	}
	now := time.Now().UTC()
	remaining := totalCount
	if duration != nil && totalCount == 0 {
		remaining = -1
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO package_instances (user_id, package_id, starts_at, expires_at, remaining_count, remaining_ai_count) VALUES (?, ?, ?, ?, ?, ?)`, item.UserID, item.PackageID, now, expiresForPackage(now, duration), remaining, aiCount)
	if err != nil {
		return Order{}, err
	}
	instanceID, err := result.LastInsertId()
	if err != nil {
		return Order{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE payment_orders SET status = ?, provider_trade_no = ?, package_instance_id = ?, paid_at = ? WHERE id = ?`, OrderPaid, notification.ProviderTradeNo, instanceID, now, item.ID); err != nil {
		return Order{}, err
	}
	if item.CouponID != 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE coupons SET reserved_count = reserved_count - 1, used_count = used_count + 1 WHERE id = ? AND reserved_count > 0`, item.CouponID); err != nil {
			return Order{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE coupon_reservations SET status = 'used' WHERE order_id = ? AND status = 'reserved'`, item.ID); err != nil {
			return Order{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}
	item.Status = OrderPaid
	item.ProviderTradeNo = notification.ProviderTradeNo
	item.PackageInstanceID = uint64(instanceID)
	item.PaidAt = &now
	return item, nil
}

func expiresForPackage(start time.Time, duration *int64) *time.Time {
	if duration == nil {
		return nil
	}
	expires := start.Add(time.Duration(*duration) * time.Second)
	return &expires
}

func (s *Store) CloseExpired(ctx context.Context, now time.Time) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id, COALESCE(coupon_id, 0) FROM payment_orders WHERE status = ? AND expires_at <= ? FOR UPDATE`, OrderPending, now)
	if err != nil {
		return 0, err
	}
	type expiredOrder struct{ id, couponID uint64 }
	expired := make([]expiredOrder, 0)
	for rows.Next() {
		var item expiredOrder
		if err := rows.Scan(&item.id, &item.couponID); err != nil {
			rows.Close()
			return 0, err
		}
		expired = append(expired, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	for _, item := range expired {
		if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status = ?, closed_at = ? WHERE id = ?`, OrderClosed, now, item.id); err != nil {
			return 0, err
		}
		if item.couponID != 0 {
			if _, err := tx.ExecContext(ctx, `UPDATE coupons SET reserved_count = reserved_count - 1 WHERE id = ? AND reserved_count > 0`, item.couponID); err != nil {
				return 0, err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE coupon_reservations SET status = 'released' WHERE order_id = ? AND status = 'reserved'`, item.id); err != nil {
				return 0, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int64(len(expired)), nil
}

func (s *Store) RecordRefund(ctx context.Context, orderNo string, amount int, reason string, refundNo string) (Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()
	var item Order
	err = tx.QueryRowContext(ctx, `SELECT o.id, o.order_no, o.user_id, o.package_id, p.name, o.provider, COALESCE(o.coupon_id, 0), o.coupon_code, o.amount_cents, o.payable_cents, o.discount_cents, o.refunded_cents, o.status, o.provider_trade_no, o.package_instance_id, o.expires_at, o.paid_at, o.closed_at, o.created_at FROM payment_orders o JOIN packages p ON p.id = o.package_id WHERE o.order_no = ? FOR UPDATE`, orderNo).
		Scan(&item.ID, &item.OrderNo, &item.UserID, &item.PackageID, &item.PackageName, &item.Provider, &item.CouponID, &item.CouponCode, &item.AmountCents, &item.PayableCents, &item.DiscountCents, &item.RefundedCents, &item.Status, &item.ProviderTradeNo, &item.PackageInstanceID, &item.ExpiresAt, &item.PaidAt, &item.ClosedAt, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, err
	}
	if item.Status != OrderPaid && item.Status != OrderPartialRefunded {
		return Order{}, errors.New("only paid orders can be refunded")
	}
	if amount <= 0 || item.RefundedCents+amount > item.PayableCents {
		return Order{}, errors.New("invalid refund amount")
	}
	newRefunded := item.RefundedCents + amount
	status := OrderPartialRefunded
	if newRefunded == item.PayableCents {
		status = OrderRefunded
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO payment_refunds (refund_no, order_id, amount_cents, reason, status) VALUES (?, ?, ?, ?, 'success')`, refundNo, item.ID, amount, reason); err != nil {
		return Order{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE payment_orders SET refunded_cents = ?, status = ? WHERE id = ?`, newRefunded, status, item.ID); err != nil {
		return Order{}, err
	}
	if newRefunded == item.PayableCents && item.PackageInstanceID != 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE package_instances SET status = 0 WHERE id = ?`, item.PackageInstanceID); err != nil {
			return Order{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}
	item.RefundedCents = newRefunded
	item.Status = status
	return item, nil
}
