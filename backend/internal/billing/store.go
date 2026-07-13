package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNoQuota = errors.New("no available package quota")

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) CreatePackage(ctx context.Context, input CreatePackageInput) (Package, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO packages (name, package_type, duration_seconds, total_count, ai_count, price_cents, limit_count, is_trial, is_free) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Type, input.DurationSeconds, input.TotalCount, input.AICount, input.PriceCents, input.LimitCount, input.IsTrial, input.IsFree)
	if err != nil {
		return Package{}, fmt.Errorf("create package: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Package{}, err
	}
	return s.GetPackage(ctx, uint64(id))
}

func (s *Store) GetPackage(ctx context.Context, id uint64) (Package, error) {
	var item Package
	err := s.db.QueryRowContext(ctx, `SELECT id, name, package_type, duration_seconds, total_count, ai_count, price_cents, status, limit_count, is_trial, is_free, created_at FROM packages WHERE id = ?`, id).
		Scan(&item.ID, &item.Name, &item.Type, &item.DurationSeconds, &item.TotalCount, &item.AICount, &item.PriceCents, &item.Status, &item.LimitCount, &item.IsTrial, &item.IsFree, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Package{}, ErrNoQuota
	}
	return item, err
}

func (s *Store) ListAvailablePackages(ctx context.Context) ([]Package, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, package_type, duration_seconds, total_count, ai_count, price_cents, status, limit_count, is_trial, is_free, created_at FROM packages WHERE status = 1 ORDER BY price_cents ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Package, 0)
	for rows.Next() {
		var item Package
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.DurationSeconds, &item.TotalCount, &item.AICount, &item.PriceCents, &item.Status, &item.LimitCount, &item.IsTrial, &item.IsFree, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListPackages(ctx context.Context) ([]Package, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, package_type, duration_seconds, total_count, ai_count, price_cents, status, limit_count, is_trial, is_free, created_at FROM packages ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Package, 0)
	for rows.Next() {
		var item Package
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.DurationSeconds, &item.TotalCount, &item.AICount, &item.PriceCents, &item.Status, &item.LimitCount, &item.IsTrial, &item.IsFree, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpdatePackageStatus(ctx context.Context, id uint64, status int) error {
	if _, err := s.GetPackage(ctx, id); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE packages SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	_ = result
	return nil
}

func (s *Store) UpdatePackage(ctx context.Context, id uint64, input UpdatePackageInput) (Package, error) {
	if _, err := s.GetPackage(ctx, id); err != nil {
		return Package{}, err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE packages SET name = ?, package_type = ?, duration_seconds = ?, total_count = ?, ai_count = ?, price_cents = ?, limit_count = ?, is_trial = ?, is_free = ? WHERE id = ?`, input.Name, input.Type, input.DurationSeconds, input.TotalCount, input.AICount, input.PriceCents, input.LimitCount, input.IsTrial, input.IsFree, id)
	if err != nil {
		return Package{}, err
	}
	_ = result
	return s.GetPackage(ctx, id)
}

func (s *Store) CreateCoupon(ctx context.Context, input CreateCouponInput) (Coupon, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO coupons (code, discount_type, discount_value, total_limit, expires_at) VALUES (?, ?, ?, ?, ?)`, input.Code, input.DiscountType, input.DiscountValue, input.TotalLimit, input.ExpiresAt)
	if err != nil {
		return Coupon{}, fmt.Errorf("create coupon: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Coupon{}, err
	}
	return s.GetCoupon(ctx, uint64(id))
}

func (s *Store) GetCoupon(ctx context.Context, id uint64) (Coupon, error) {
	var item Coupon
	err := s.db.QueryRowContext(ctx, `SELECT id, code, discount_type, discount_value, total_limit, used_count, reserved_count, expires_at, status FROM coupons WHERE id = ?`, id).
		Scan(&item.ID, &item.Code, &item.DiscountType, &item.DiscountValue, &item.TotalLimit, &item.UsedCount, &item.ReservedCount, &item.ExpiresAt, &item.Status)
	return item, err
}

func (s *Store) ListCoupons(ctx context.Context) ([]Coupon, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, code, discount_type, discount_value, total_limit, used_count, reserved_count, expires_at, status FROM coupons ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Coupon, 0)
	for rows.Next() {
		var item Coupon
		if err := rows.Scan(&item.ID, &item.Code, &item.DiscountType, &item.DiscountValue, &item.TotalLimit, &item.UsedCount, &item.ReservedCount, &item.ExpiresAt, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpdateCouponStatus(ctx context.Context, id uint64, status int) error {
	if _, err := s.GetCoupon(ctx, id); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE coupons SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	_ = result
	return nil
}

func (s *Store) GrantPackage(ctx context.Context, userID, packageID uint64) (PackageInstance, error) {
	item, err := s.GetPackage(ctx, packageID)
	if err != nil {
		return PackageInstance{}, err
	}
	if item.Status != 1 {
		return PackageInstance{}, errors.New("package is not available")
	}
	now := time.Now().UTC()
	var expiresAt *time.Time
	if item.DurationSeconds != nil {
		expires := now.Add(time.Duration(*item.DurationSeconds) * time.Second)
		expiresAt = &expires
	}
	remaining := item.TotalCount
	if item.Type == PackageTime {
		remaining = -1
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO package_instances (user_id, package_id, starts_at, expires_at, remaining_count, remaining_ai_count) VALUES (?, ?, ?, ?, ?, ?)`, userID, packageID, now, expiresAt, remaining, item.AICount)
	if err != nil {
		return PackageInstance{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return PackageInstance{}, err
	}
	return s.GetInstance(ctx, uint64(id))
}

func (s *Store) GetInstance(ctx context.Context, id uint64) (PackageInstance, error) {
	var item PackageInstance
	err := s.db.QueryRowContext(ctx, `SELECT i.id, i.package_id, p.name, p.package_type, i.starts_at, i.expires_at, i.remaining_count, i.remaining_ai_count, i.status FROM package_instances i JOIN packages p ON p.id = i.package_id WHERE i.id = ?`, id).
		Scan(&item.ID, &item.PackageID, &item.PackageName, &item.PackageType, &item.StartsAt, &item.ExpiresAt, &item.RemainingCount, &item.RemainingAICount, &item.Status)
	return item, err
}

func (s *Store) ListInstances(ctx context.Context, userID uint64) ([]PackageInstance, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT i.id, i.package_id, p.name, p.package_type, i.starts_at, i.expires_at, i.remaining_count, i.remaining_ai_count, i.status FROM package_instances i JOIN packages p ON p.id = i.package_id WHERE i.user_id = ? ORDER BY (i.expires_at IS NULL), i.expires_at ASC, i.id ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PackageInstance, 0)
	for rows.Next() {
		var item PackageInstance
		if err := rows.Scan(&item.ID, &item.PackageID, &item.PackageName, &item.PackageType, &item.StartsAt, &item.ExpiresAt, &item.RemainingCount, &item.RemainingAICount, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) Consume(ctx context.Context, userID, packageID uint64, kind, requestID, endpoint string, amount int) (PackageInstance, error) {
	if amount <= 0 {
		amount = 1
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PackageInstance{}, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	query := `SELECT i.id, i.package_id, p.name, p.package_type, i.starts_at, i.expires_at, i.remaining_count, i.remaining_ai_count, i.status FROM package_instances i JOIN packages p ON p.id = i.package_id WHERE i.user_id = ? AND i.status = 1 AND i.starts_at <= ? AND (i.expires_at IS NULL OR i.expires_at > ?)`
	args := []any{userID, now, now}
	if packageID != 0 {
		query += ` AND i.package_id = ?`
		args = append(args, packageID)
	}
	query += ` ORDER BY (i.expires_at IS NULL), i.expires_at ASC, i.id ASC FOR UPDATE`
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return PackageInstance{}, err
	}
	var selected PackageInstance
	for rows.Next() {
		var item PackageInstance
		if err := rows.Scan(&item.ID, &item.PackageID, &item.PackageName, &item.PackageType, &item.StartsAt, &item.ExpiresAt, &item.RemainingCount, &item.RemainingAICount, &item.Status); err != nil {
			rows.Close()
			return PackageInstance{}, err
		}
		available := item.RemainingCount == -1
		if kind == UsageAI {
			available = item.RemainingAICount >= amount
		} else if item.RemainingCount >= amount {
			available = true
		}
		if available {
			selected = item
			break
		}
	}
	rows.Close()
	if selected.ID == 0 {
		return PackageInstance{}, ErrNoQuota
	}
	if kind == UsageAI {
		_, err = tx.ExecContext(ctx, `UPDATE package_instances SET remaining_ai_count = remaining_ai_count - ? WHERE id = ? AND remaining_ai_count >= ?`, amount, selected.ID, amount)
	} else if selected.RemainingCount > 0 {
		_, err = tx.ExecContext(ctx, `UPDATE package_instances SET remaining_count = remaining_count - ? WHERE id = ? AND remaining_count >= ?`, amount, selected.ID, amount)
	}
	if err != nil {
		return PackageInstance{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO package_consumptions (instance_id, user_id, kind, amount, request_id, endpoint) VALUES (?, ?, ?, ?, ?, ?)`, selected.ID, userID, kind, amount, requestID, endpoint); err != nil {
		return PackageInstance{}, err
	}
	if err = tx.Commit(); err != nil {
		return PackageInstance{}, err
	}
	if selected.RemainingCount > 0 {
		selected.RemainingCount--
	}
	if kind == UsageAI {
		selected.RemainingAICount -= amount
	} else if selected.RemainingCount > 0 {
		selected.RemainingCount -= amount
	}
	return selected, nil
}

// HasAIQuota performs a non-mutating preflight before an external AI request.
// Consume remains the authoritative transactional deduction after a successful response.
func (s *Store) HasAIQuota(ctx context.Context, userID, packageID uint64) (bool, error) {
	now := time.Now().UTC()
	query := `SELECT EXISTS(SELECT 1 FROM package_instances i WHERE i.user_id = ? AND i.status = 1 AND i.starts_at <= ? AND (i.expires_at IS NULL OR i.expires_at > ?) AND i.remaining_ai_count >= 1`
	args := []any{userID, now, now}
	if packageID != 0 {
		query += ` AND i.package_id = ?`
		args = append(args, packageID)
	}
	query += `)`
	var exists bool
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
