package ocs

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (s *Store) CreateSource(ctx context.Context, input SourceInput) (Source, error) {
	headers, err := json.Marshal(input.Headers)
	if err != nil {
		return Source{}, err
	}
	data, err := json.Marshal(input.Data)
	if err != nil {
		return Source{}, err
	}
	method := input.Method
	if method == "" {
		method = "GET"
	}
	enabled := 0
	if input.Enabled {
		enabled = 1
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO ocs_sources (name, homepage, url, method, headers_json, data_json, success_path, success_value, question_path, answer_path, priority, enabled) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, input.Name, input.Homepage, input.URL, method, headers, data, input.SuccessPath, input.SuccessValue, input.QuestionPath, input.AnswerPath, input.Priority, enabled)
	if err != nil {
		return Source{}, fmt.Errorf("create OCS source: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Source{}, err
	}
	return s.GetSource(ctx, uint64(id))
}

func (s *Store) UpdateSource(ctx context.Context, id uint64, input SourceInput) error {
	headers, err := json.Marshal(input.Headers)
	if err != nil {
		return err
	}
	data, err := json.Marshal(input.Data)
	if err != nil {
		return err
	}
	enabled := 0
	if input.Enabled {
		enabled = 1
	}
	result, err := s.db.ExecContext(ctx, `UPDATE ocs_sources SET name = ?, homepage = ?, url = ?, method = ?, headers_json = ?, data_json = ?, success_path = ?, success_value = ?, question_path = ?, answer_path = ?, priority = ?, enabled = ? WHERE id = ?`, input.Name, input.Homepage, input.URL, input.Method, headers, data, input.SuccessPath, input.SuccessValue, input.QuestionPath, input.AnswerPath, input.Priority, enabled, id)
	if err != nil {
		return fmt.Errorf("update OCS source: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateSourceStatus(ctx context.Context, id uint64, status int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE ocs_sources SET enabled = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("update OCS source status: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetSource(ctx context.Context, id uint64) (Source, error) {
	var item Source
	var headers, data string
	err := s.db.QueryRowContext(ctx, `SELECT id, name, homepage, url, method, headers_json, data_json, success_path, success_value, question_path, answer_path, priority, enabled, created_at FROM ocs_sources WHERE id = ?`, id).
		Scan(&item.ID, &item.Name, &item.Homepage, &item.URL, &item.Method, &headers, &data, &item.SuccessPath, &item.SuccessValue, &item.QuestionPath, &item.AnswerPath, &item.Priority, &item.Enabled, &item.CreatedAt)
	if err != nil {
		return Source{}, err
	}
	if err := json.Unmarshal([]byte(headers), &item.Headers); err != nil {
		return Source{}, err
	}
	if err := json.Unmarshal([]byte(data), &item.Data); err != nil {
		return Source{}, err
	}
	return item, nil
}

func (s *Store) ListSources(ctx context.Context) ([]Source, error) {
	return s.listSources(ctx, ``)
}

func (s *Store) ListEnabledSources(ctx context.Context) ([]Source, error) {
	return s.listSources(ctx, ` WHERE enabled = 1`)
}

func (s *Store) listSources(ctx context.Context, condition string) ([]Source, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, homepage, url, method, headers_json, data_json, success_path, success_value, question_path, answer_path, priority, enabled, created_at FROM ocs_sources`+condition+` ORDER BY priority ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Source, 0)
	for rows.Next() {
		var item Source
		var headers, data string
		if err := rows.Scan(&item.ID, &item.Name, &item.Homepage, &item.URL, &item.Method, &headers, &data, &item.SuccessPath, &item.SuccessValue, &item.QuestionPath, &item.AnswerPath, &item.Priority, &item.Enabled, &item.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(headers), &item.Headers); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(data), &item.Data); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
