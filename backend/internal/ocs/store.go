package ocs

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) SaveConfig(ctx context.Context, userID uint64, configJSON string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO answerer_configs (user_id, name, config_json) VALUES (?, 'TIKU-ZONG', ?) ON DUPLICATE KEY UPDATE config_json = VALUES(config_json), updated_at = CURRENT_TIMESTAMP(6)`, userID, configJSON)
	if err != nil {
		return fmt.Errorf("save OCS config: %w", err)
	}
	return nil
}
