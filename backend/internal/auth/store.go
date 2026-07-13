package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrInviteInvalid = errors.New("invitation code is invalid or unavailable")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO users (email, password_hash) VALUES (?, ?)`, email, passwordHash)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("read user id: %w", err)
	}
	return s.GetUserByID(ctx, uint64(id))
}

func (s *Store) CreateUserWithInvite(ctx context.Context, email, passwordHash, inviteCode string) (User, error) {
	inviteCode = strings.ToUpper(strings.TrimSpace(inviteCode))
	if inviteCode == "" {
		return s.CreateUser(ctx, email, passwordHash)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	var inviteID uint64
	var maxUses, usedCount, status int
	var expiresAt *time.Time
	err = tx.QueryRowContext(ctx, `SELECT id, max_uses, used_count, status, expires_at FROM invitation_codes WHERE code = ? FOR UPDATE`, inviteCode).
		Scan(&inviteID, &maxUses, &usedCount, &status, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrInviteInvalid
	}
	if err != nil {
		return User{}, err
	}
	if status != 1 || (maxUses > 0 && usedCount >= maxUses) || expiresAt != nil && !time.Now().UTC().Before(*expiresAt) {
		return User{}, ErrInviteInvalid
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO users (email, password_hash) VALUES (?, ?)`, email, passwordHash)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE invitation_codes SET used_count = used_count + 1 WHERE id = ?`, inviteID); err != nil {
		return User{}, err
	}
	if err = tx.Commit(); err != nil {
		return User{}, err
	}
	return s.GetUserByID(ctx, uint64(userID))
}

func (s *Store) CreateInvite(ctx context.Context, actorID uint64, input CreateInviteInput) (InviteView, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if code == "" || len(code) > 64 || input.MaxUses < 0 {
		return InviteView{}, ErrInvalidInput
	}
	status := input.Status
	if status != 0 && status != 1 {
		status = 1
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO invitation_codes (code, max_uses, expires_at, status, created_by) VALUES (?, ?, ?, ?, ?)`, code, input.MaxUses, input.ExpiresAt, status, actorID)
	if err != nil {
		return InviteView{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return InviteView{}, err
	}
	return s.GetInvite(ctx, uint64(id))
}

func (s *Store) ListInvites(ctx context.Context) ([]InviteView, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, code, max_uses, used_count, status, expires_at, COALESCE(created_by, 0), created_at FROM invitation_codes ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]InviteView, 0)
	for rows.Next() {
		var item InviteView
		if err := rows.Scan(&item.ID, &item.Code, &item.MaxUses, &item.UsedCount, &item.Status, &item.ExpiresAt, &item.CreatedBy, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetInvite(ctx context.Context, id uint64) (InviteView, error) {
	var item InviteView
	err := s.db.QueryRowContext(ctx, `SELECT id, code, max_uses, used_count, status, expires_at, COALESCE(created_by, 0), created_at FROM invitation_codes WHERE id = ?`, id).
		Scan(&item.ID, &item.Code, &item.MaxUses, &item.UsedCount, &item.Status, &item.ExpiresAt, &item.CreatedBy, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return InviteView{}, ErrNotFound
	}
	return item, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, string, error) {
	var user User
	var passwordHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, role, status, failed_login_count, locked_until, last_login_at, created_at
		FROM users WHERE email = ?`, email).
		Scan(&user.ID, &user.Email, &passwordHash, &user.Role, &user.Status, &user.FailedLoginCount, &user.LockedUntil, &user.LastLoginAt, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	if err != nil {
		return User{}, "", fmt.Errorf("get user by email: %w", err)
	}
	return user, passwordHash, nil
}

func (s *Store) UpdatePassword(ctx context.Context, userID uint64, passwordHash string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, userID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeRefreshTokens invalidates every renewable session for a user. Access
// tokens remain valid only until their normal short-lived JWT expiry.
func (s *Store) RevokeRefreshTokens(ctx context.Context, userID uint64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`, time.Now().UTC(), userID)
	return err
}

func (s *Store) GetUserByID(ctx context.Context, id uint64) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, role, status, failed_login_count, locked_until, last_login_at, created_at
		FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Email, &user.Role, &user.Status, &user.FailedLoginCount, &user.LockedUntil, &user.LastLoginAt, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *Store) GetUserByIDWithPassword(ctx context.Context, id uint64) (User, string, error) {
	var user User
	var passwordHash string
	err := s.db.QueryRowContext(ctx, `SELECT id, email, password_hash, role, status, failed_login_count, locked_until, last_login_at, created_at FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Email, &passwordHash, &user.Role, &user.Status, &user.FailedLoginCount, &user.LockedUntil, &user.LastLoginAt, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	return user, passwordHash, err
}

func (s *Store) SetRole(ctx context.Context, userID uint64, role string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, userID)
	return err
}

func (s *Store) ListUsers(ctx context.Context, search string, status, page, pageSize int) (AdminUserPage, error) {
	page, pageSize = normalizePage(page, pageSize)
	search = strings.TrimSpace(search)
	where := "WHERE 1 = 1"
	args := make([]any, 0, 4)
	if search != "" {
		where += " AND email LIKE ?"
		args = append(args, "%"+search+"%")
	}
	if status == 0 || status == 1 {
		where += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		return AdminUserPage{}, err
	}
	args = append(args, (page-1)*pageSize, pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT u.id, u.email, u.role, u.status, u.failed_login_count,
		u.locked_until, u.last_login_at, u.created_at, COALESCE(k.key_prefix, '')
		FROM users u LEFT JOIN user_api_keys k ON k.user_id = u.id AND k.revoked_at IS NULL `+where+`
		ORDER BY u.id DESC LIMIT ?, ?`, args...)
	if err != nil {
		return AdminUserPage{}, err
	}
	defer rows.Close()
	items := make([]AdminUserView, 0)
	for rows.Next() {
		var item AdminUserView
		if err := rows.Scan(&item.ID, &item.Email, &item.Role, &item.Status, &item.FailedLoginCount,
			&item.LockedUntil, &item.LastLoginAt, &item.CreatedAt, &item.APIKeyPrefix); err != nil {
			return AdminUserPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return AdminUserPage{}, err
	}
	return AdminUserPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Store) UpdateStatus(ctx context.Context, userID uint64, status int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE users SET status = ? WHERE id = ?`, status, userID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateRole(ctx context.Context, userID uint64, role string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, userID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
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

func (s *Store) RecordLoginFailure(ctx context.Context, id uint64, lockedUntil *time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET failed_login_count = failed_login_count + 1, locked_until = ? WHERE id = ?`, lockedUntil, id)
	return err
}

func (s *Store) RecordLoginSuccess(ctx context.Context, id uint64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET failed_login_count = 0, locked_until = NULL, last_login_at = ? WHERE id = ?`, time.Now().UTC(), id)
	return err
}

func (s *Store) SaveRefreshToken(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES (UUID(), ?, ?, ?)`, userID, tokenHash, expiresAt)
	return err
}

func (s *Store) ConsumeRefreshToken(ctx context.Context, tokenHash string) (uint64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var id uint64
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, `SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = ? AND revoked_at IS NULL FOR UPDATE`, tokenHash).Scan(&id, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if time.Now().UTC().After(expiresAt) {
		return 0, ErrNotFound
	}
	if _, err = tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ?`, time.Now().UTC(), tokenHash); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) GetAPIKey(ctx context.Context, userID uint64) (APIKeyView, error) {
	var view APIKeyView
	err := s.db.QueryRowContext(ctx, `SELECT key_prefix, last_used_at, created_at FROM user_api_keys WHERE user_id = ? AND revoked_at IS NULL`, userID).
		Scan(&view.Prefix, &view.LastUsedAt, &view.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return APIKeyView{}, ErrNotFound
	}
	if err != nil {
		return APIKeyView{}, err
	}
	view.Masked = maskAPIKey(view.Prefix + "************")
	return view, nil
}

func (s *Store) CreateAPIKey(ctx context.Context, userID uint64) (string, APIKeyView, error) {
	plain, hash, prefix, err := newAPIKey()
	if err != nil {
		return "", APIKeyView{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO user_api_keys (user_id, key_prefix, key_hash) VALUES (?, ?, ?)`, userID, prefix, hash)
	if err != nil {
		return "", APIKeyView{}, fmt.Errorf("create api key: %w", err)
	}
	view := APIKeyView{Prefix: prefix, Masked: maskAPIKey(plain), CreatedAt: time.Now()}
	return plain, view, nil
}

func (s *Store) RotateAPIKey(ctx context.Context, userID uint64) (string, APIKeyView, error) {
	plain, hash, prefix, err := newAPIKey()
	if err != nil {
		return "", APIKeyView{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", APIKeyView{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE user_api_keys SET key_prefix = ?, key_hash = ?, revoked_at = NULL, last_used_at = NULL WHERE user_id = ?`, prefix, hash, userID)
	if err != nil {
		return "", APIKeyView{}, fmt.Errorf("rotate api key: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return "", APIKeyView{}, err
	}
	if affected == 0 {
		if _, err = tx.ExecContext(ctx, `INSERT INTO user_api_keys (user_id, key_prefix, key_hash) VALUES (?, ?, ?)`, userID, prefix, hash); err != nil {
			return "", APIKeyView{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", APIKeyView{}, err
	}
	return plain, APIKeyView{Prefix: prefix, Masked: maskAPIKey(plain), CreatedAt: time.Now()}, nil
}

func (s *Store) ResolveAPIKey(ctx context.Context, plain string) (User, uint64, error) {
	var user User
	var keyID uint64
	err := s.db.QueryRowContext(ctx, `
		SELECT k.id, u.id, u.email, u.role, u.status, u.failed_login_count, u.locked_until, u.last_login_at, u.created_at
		FROM user_api_keys k JOIN users u ON u.id = k.user_id
		WHERE k.key_hash = ? AND k.revoked_at IS NULL`, hashToken(plain)).
		Scan(&keyID, &user.ID, &user.Email, &user.Role, &user.Status, &user.FailedLoginCount, &user.LockedUntil, &user.LastLoginAt, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, 0, ErrNotFound
	}
	if err != nil {
		return User{}, 0, err
	}
	if user.Status != 1 {
		return User{}, 0, ErrNotFound
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE user_api_keys SET last_used_at = ? WHERE id = ?`, time.Now().UTC(), keyID)
	return user, keyID, nil
}
