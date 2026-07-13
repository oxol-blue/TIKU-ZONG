package ai

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oxol-blue/TIKU-ZONG/backend/internal/secret"
)

var ErrNotFound = errors.New("ai record not found")

type Store struct {
	db     *sql.DB
	secret string
}

func NewStore(db *sql.DB, masterSecret string) *Store { return &Store{db: db, secret: masterSecret} }

func (s *Store) CreateProvider(ctx context.Context, input CreateProviderInput) (uint64, error) {
	ciphertext, err := secret.Encrypt(input.APIKey, s.secret)
	if err != nil {
		return 0, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO ai_providers (name, base_url, api_key_ciphertext) VALUES (?, ?, ?)`, input.Name, input.BaseURL, ciphertext)
	if err != nil {
		return 0, fmt.Errorf("create ai provider: %w", err)
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (s *Store) CreateModel(ctx context.Context, input CreateModelInput) (uint64, error) {
	if input.Priority == 0 {
		input.Priority = 100
	}
	if input.TimeoutSeconds == 0 {
		input.TimeoutSeconds = 30
	}
	if input.AIChargeCount == 0 {
		input.AIChargeCount = 1
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO ai_models (provider_id, name, priority, timeout_seconds, ai_charge_count) VALUES (?, ?, ?, ?, ?)`, input.ProviderID, input.Name, input.Priority, input.TimeoutSeconds, input.AIChargeCount)
	if err != nil {
		return 0, fmt.Errorf("create ai model: %w", err)
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (s *Store) ListModels(ctx context.Context) ([]Model, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.id, m.provider_id, p.name, p.base_url, p.api_key_ciphertext, m.name, m.priority, m.timeout_seconds, m.ai_charge_count FROM ai_models m JOIN ai_providers p ON p.id = m.provider_id WHERE m.enabled = 1 AND p.enabled = 1 ORDER BY m.priority ASC, m.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Model, 0)
	for rows.Next() {
		var item Model
		if err := rows.Scan(&item.ID, &item.ProviderID, &item.ProviderName, &item.BaseURL, &item.EncryptedKey, &item.Name, &item.Priority, &item.TimeoutSeconds, &item.AIChargeCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetCached(ctx context.Context, hash string) (Answer, error) {
	var item Answer
	var elapsedMicros int64
	err := s.db.QueryRowContext(ctx, `SELECT question_hash, question_text, question_type, answer_text, raw_response, provider_name, model_name, token_count, charge_count, elapsed_micros FROM question_ai WHERE question_hash = ? AND status = 1`, hash).
		Scan(&item.QuestionHash, &item.Question, &item.Type, &item.Text, &item.RawResponse, &item.Provider, &item.Model, &item.TokenCount, &item.ChargeCount, &elapsedMicros)
	if errors.Is(err, sql.ErrNoRows) {
		return Answer{}, ErrNotFound
	}
	item.Elapsed = time.Duration(elapsedMicros) * time.Microsecond
	if item.ChargeCount <= 0 {
		item.ChargeCount = 1
	}
	return item, err
}

func (s *Store) Save(ctx context.Context, item Answer, prompt string) error {
	if item.ChargeCount <= 0 {
		item.ChargeCount = 1
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO question_ai (question_hash, question_text, question_type, answer_text, prompt_text, raw_response, provider_name, model_name, token_count, charge_count, elapsed_micros) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE answer_text = VALUES(answer_text), raw_response = VALUES(raw_response), prompt_text = VALUES(prompt_text), provider_name = VALUES(provider_name), model_name = VALUES(model_name), token_count = VALUES(token_count), charge_count = VALUES(charge_count), elapsed_micros = VALUES(elapsed_micros), status = 1, updated_at = CURRENT_TIMESTAMP(6)`, item.QuestionHash, item.Question, item.Type, item.Text, prompt, item.RawResponse, item.Provider, item.Model, item.TokenCount, item.ChargeCount, item.Elapsed.Microseconds())
	return err
}

func (s *Store) ListAnswers(ctx context.Context, search, provider, model string, status, page, pageSize int) (AnswerPage, error) {
	page, pageSize = normalizePage(page, pageSize)
	where := "WHERE 1 = 1"
	args := make([]any, 0, 8)
	if strings.TrimSpace(search) != "" {
		where += " AND question_text LIKE ?"
		args = append(args, "%"+strings.TrimSpace(search)+"%")
	}
	if strings.TrimSpace(provider) != "" {
		where += " AND provider_name = ?"
		args = append(args, strings.TrimSpace(provider))
	}
	if strings.TrimSpace(model) != "" {
		where += " AND model_name = ?"
		args = append(args, strings.TrimSpace(model))
	}
	if status == 0 || status == 1 {
		where += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM question_ai "+where, args...).Scan(&total); err != nil {
		return AnswerPage{}, err
	}
	args = append(args, (page-1)*pageSize, pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT id, question_hash, question_text, question_type, answer_text,
		prompt_text, raw_response, provider_name, model_name, token_count, elapsed_micros, status, created_at, updated_at
		FROM question_ai `+where+` ORDER BY id DESC LIMIT ?, ?`, args...)
	if err != nil {
		return AnswerPage{}, err
	}
	defer rows.Close()
	items := make([]AdminAnswer, 0)
	for rows.Next() {
		var item AdminAnswer
		if err := rows.Scan(&item.ID, &item.QuestionHash, &item.Question, &item.Type, &item.Text, &item.Prompt,
			&item.RawResponse, &item.Provider, &item.Model, &item.TokenCount, &item.Elapsed, &item.Status,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return AnswerPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return AnswerPage{}, err
	}
	return AnswerPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Store) GetAnswer(ctx context.Context, id uint64) (AdminAnswer, error) {
	var item AdminAnswer
	err := s.db.QueryRowContext(ctx, `SELECT id, question_hash, question_text, question_type, answer_text,
		prompt_text, raw_response, provider_name, model_name, token_count, elapsed_micros, status, created_at, updated_at
		FROM question_ai WHERE id = ?`, id).Scan(&item.ID, &item.QuestionHash, &item.Question, &item.Type, &item.Text,
		&item.Prompt, &item.RawResponse, &item.Provider, &item.Model, &item.TokenCount, &item.Elapsed, &item.Status,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminAnswer{}, ErrNotFound
	}
	return item, err
}

func (s *Store) UpdateAnswerStatus(ctx context.Context, id uint64, status int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE question_ai SET status = ? WHERE id = ?`, status, id)
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

func (s *Store) DecryptKey(value string) (string, error) { return secret.Decrypt(value, s.secret) }
