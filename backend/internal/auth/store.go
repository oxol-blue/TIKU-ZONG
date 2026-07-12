package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("not found")

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

func (s *Store) SetRole(ctx context.Context, userID uint64, role string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, userID)
	return err
}

func (s *Store) RecordLoginFailure(ctx context.Context, id uint64, lockedUntil *time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET failed_login_count = failed_login_count + 1, locked_until = ? WHERE id = ?`, lockedUntil, id)
	return err
}

func (s *Store) RecordLoginSuccess(ctx context.Context, id uint64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET failed_login_count = 0, locked_until = NULL, last_login_at = ? WHERE id = ?`, time.Now(), id)
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
	if time.Now().After(expiresAt) {
		return 0, ErrNotFound
	}
	if _, err = tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ?`, time.Now(), tokenHash); err != nil {
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
