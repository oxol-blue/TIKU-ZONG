package settings

import (
	"context"
	"database/sql"
	"strings"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT setting_value FROM system_settings WHERE setting_key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *Store) GetMany(ctx context.Context, keys []string) (map[string]string, error) {
	items := make(map[string]string, len(keys))
	for _, key := range keys {
		value, err := s.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		items[key] = value
	}
	return items, nil
}

func (s *Store) Put(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		public := 1
		if _, err := s.db.ExecContext(ctx, `INSERT INTO system_settings (setting_key, setting_value, is_public) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value), is_public = VALUES(is_public), updated_at = CURRENT_TIMESTAMP(6)`, key, strings.TrimSpace(value), public); err != nil {
			return err
		}
	}
	return nil
}
