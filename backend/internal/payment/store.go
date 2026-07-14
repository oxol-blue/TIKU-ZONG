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
	input.SecretKey = strings.TrimSpace(input.SecretKey)
	var ciphertext string
	if input.SecretKey == "" {
		err := s.db.QueryRowContext(ctx, `SELECT secret_ciphertext FROM payment_gateways WHERE provider = ?`, input.Provider).Scan(&ciphertext)
		if errors.Is(err, sql.ErrNoRows) {
			return Gateway{}, errors.New("secretKey is required when creating a payment gateway")
		}
		if err != nil {
			return Gateway{}, err
		}
	} else {
		var err error
		ciphertext, err = secret.Encrypt(input.SecretKey, s.secret)
		if err != nil {
			return Gateway{}, err
		}
	}
	enabled := 0
	if input.Enabled {
		enabled = 1
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO payment_gateways (provider, name, base_url, merchant_id, secret_ciphertext, enabled) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name = VALUES(name), base_url = VALUES(base_url), merchant_id = VALUES(merchant_id), secret_ciphertext = VALUES(secret_ciphertext), enabled = VALUES(enabled), updated_at = CURRENT_TIMESTAMP(6)`, input.Provider, input.Name, input.BaseURL, input.MerchantID, ciphertext, enabled)
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

func (s *Store) GetOrderForUser(ctx context.Context, orderNo string, userID uint64) (Order, error) {
	item, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return Order{}, err
	}
	if item.UserID != userID {
		return Order{}, ErrOrderNotFound
	}
	return item, nil
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

func (s *Store) ListAdminOrders(ctx context.Context, search, status string, page, pageSize int) (OrderPage, error) {
	page, pageSize = normalizePage(page, pageSize)
	where := "WHERE 1 = 1"
	args := make([]any, 0, 4)
	if strings.TrimSpace(search) != "" {
		where += " AND (o.order_no LIKE ? OR u.email LIKE ?)"
		value := "%" + strings.TrimSpace(search) + "%"
		args = append(args, value, value)
	}
	if strings.TrimSpace(status) != "" {
		where += " AND o.status = ?"
		args = append(args, strings.TrimSpace(status))
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payment_orders o JOIN users u ON u.id = o.user_id "+where, args...).Scan(&total); err != nil {
		return OrderPage{}, err
	}
	args = append(args, (page-1)*pageSize, pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT o.id, o.order_no, o.user_id, u.email, o.package_id, p.name, o.provider,
		COALESCE(o.coupon_id, 0), o.coupon_code, o.amount_cents, o.payable_cents, o.discount_cents,
		o.refunded_cents, o.status, o.provider_trade_no, o.package_instance_id, o.expires_at,
		o.paid_at, o.closed_at, o.created_at
		FROM payment_orders o JOIN users u ON u.id = o.user_id JOIN packages p ON p.id = o.package_id `+where+`
		ORDER BY o.id DESC LIMIT ?, ?`, args...)
	if err != nil {
		return OrderPage{}, err
	}
	defer rows.Close()
	items := make([]AdminOrderView, 0)
	for rows.Next() {
		var item AdminOrderView
		if err := rows.Scan(&item.ID, &item.OrderNo, &item.UserID, &item.UserEmail, &item.PackageID, &item.PackageName,
			&item.Provider, &item.CouponID, &item.CouponCode, &item.AmountCents, &item.PayableCents, &item.DiscountCents,
			&item.RefundedCents, &item.Status, &item.ProviderTradeNo, &item.PackageInstanceID, &item.ExpiresAt,
			&item.PaidAt, &item.ClosedAt, &item.CreatedAt); err != nil {
			return OrderPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return OrderPage{}, err
	}
	return OrderPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
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
	remaining := remainingForPackage(totalCount, duration)
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

func remainingForPackage(totalCount int, duration *int64) int {
	if duration != nil && totalCount == 0 {
		return -1
	}
	return totalCount
}

// RepairMissingPackageInstances safely grants packages for orders that were
// paid but lost their package-instance link due to a historic partial failure.
// Fully refunded orders are intentionally excluded: there is no entitlement to
// restore. The transaction locks each order and only updates an empty link, so
// repeated maintenance runs are idempotent.
func (s *Store) RepairMissingPackageInstances(ctx context.Context) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	type candidate struct {
		id        uint64
		userID    uint64
		packageID uint64
		startsAt  time.Time
		duration  *int64
		total     int
		aiCount   int
	}
	rows, err := tx.QueryContext(ctx, `SELECT o.id, o.user_id, o.package_id, COALESCE(o.paid_at, o.created_at), p.duration_seconds, p.total_count, p.ai_count
		FROM payment_orders o JOIN packages p ON p.id = o.package_id
		WHERE o.status IN (?, ?) AND o.package_instance_id = 0 FOR UPDATE`, OrderPaid, OrderPartialRefunded)
	if err != nil {
		return 0, err
	}
	candidates := make([]candidate, 0)
	for rows.Next() {
		var item candidate
		if err := rows.Scan(&item.id, &item.userID, &item.packageID, &item.startsAt, &item.duration, &item.total, &item.aiCount); err != nil {
			rows.Close()
			return 0, err
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	for _, item := range candidates {
		result, err := tx.ExecContext(ctx, `INSERT INTO package_instances (user_id, package_id, starts_at, expires_at, remaining_count, remaining_ai_count) VALUES (?, ?, ?, ?, ?, ?)`, item.userID, item.packageID, item.startsAt, expiresForPackage(item.startsAt, item.duration), remainingForPackage(item.total, item.duration), item.aiCount)
		if err != nil {
			return 0, err
		}
		instanceID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		updated, err := tx.ExecContext(ctx, `UPDATE payment_orders SET package_instance_id = ? WHERE id = ? AND package_instance_id = 0`, instanceID, item.id)
		if err != nil {
			return 0, err
		}
		if affected, _ := updated.RowsAffected(); affected != 1 {
			return 0, errors.New("payment order package instance was updated concurrently")
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int64(len(candidates)), nil
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
	var existingOrderNo string
	err = tx.QueryRowContext(ctx, `SELECT o.order_no FROM payment_refunds r JOIN payment_orders o ON o.id = r.order_id WHERE r.refund_no = ?`, refundNo).Scan(&existingOrderNo)
	if err == nil {
		if existingOrderNo != orderNo {
			return Order{}, errors.New("refund number belongs to another order")
		}
		if err := tx.Commit(); err != nil {
			return Order{}, err
		}
		return s.GetOrder(ctx, orderNo)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Order{}, err
	}
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

func (s *Store) ListRefunds(ctx context.Context, orderNo string) ([]Refund, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.id, r.refund_no, o.order_no, r.amount_cents, r.reason, r.status, r.created_at
		FROM payment_refunds r JOIN payment_orders o ON o.id = r.order_id WHERE o.order_no = ? ORDER BY r.id DESC`, orderNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Refund, 0)
	for rows.Next() {
		var item Refund
		if err := rows.Scan(&item.ID, &item.RefundNo, &item.OrderNo, &item.AmountCents, &item.Reason, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) Reconcile(ctx context.Context, now time.Time) ([]ReconciliationIssue, error) {
	issues := make([]ReconciliationIssue, 0)
	rows, err := s.db.QueryContext(ctx, `SELECT order_no FROM payment_orders WHERE status IN (?, ?) AND package_instance_id = 0`, OrderPaid, OrderPartialRefunded)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var orderNo string
		if err := rows.Scan(&orderNo); err != nil {
			rows.Close()
			return nil, err
		}
		issues = append(issues, ReconciliationIssue{OrderNo: orderNo, IssueType: "missing_package_instance", Detail: "paid order has no package instance"})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT order_no FROM payment_orders WHERE status = ? AND expires_at <= ?`, OrderPending, now)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var orderNo string
		if err := rows.Scan(&orderNo); err != nil {
			rows.Close()
			return nil, err
		}
		issues = append(issues, ReconciliationIssue{OrderNo: orderNo, IssueType: "expired_pending", Detail: "pending order is past its expiry time"})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT o.order_no, o.payable_cents, o.refunded_cents, COALESCE(SUM(r.amount_cents), 0)
		FROM payment_orders o LEFT JOIN payment_refunds r ON r.order_id = o.id
		WHERE o.status IN (?, ?) GROUP BY o.id, o.order_no, o.payable_cents, o.refunded_cents
		HAVING COALESCE(SUM(r.amount_cents), 0) <> o.refunded_cents OR o.refunded_cents > o.payable_cents`, OrderPartialRefunded, OrderRefunded)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var orderNo string
		var payable, recorded, refundTotal int
		if err := rows.Scan(&orderNo, &payable, &recorded, &refundTotal); err != nil {
			rows.Close()
			return nil, err
		}
		issues = append(issues, ReconciliationIssue{OrderNo: orderNo, IssueType: "refund_amount_mismatch", Detail: fmt.Sprintf("payable=%d recorded=%d refund_records=%d", payable, recorded, refundTotal)})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	return issues, nil
}
