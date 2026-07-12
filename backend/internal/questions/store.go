package questions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("question not found")

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Upsert(ctx context.Context, input ImportInput) (Question, bool, error) {
	questionText := normalizeText(input.Question)
	_, optionsHash := normalizeOptions(input.Options)
	answerText := normalizeAnswer(input.Answer, input.Options)
	questionHash := hashText(questionText)
	answerHash := hashText(answerText)
	compositeHash := hashText(strings.Join([]string{questionHash, optionsHash, answerHash}, ":"))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Question{}, false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO questions
		(question_text, normalized_text, question_hash, options_hash, answer_hash, composite_hash, question_type, platform, subject, source, collected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, questionText, questionText, questionHash, optionsHash, answerHash, compositeHash,
		defaultType(input.Type), strings.TrimSpace(input.Platform), strings.TrimSpace(input.Subject), strings.TrimSpace(input.Source), input.CollectedAt)
	if err != nil {
		return Question{}, false, fmt.Errorf("insert question: %w", err)
	}
	created := false
	if affected, _ := result.RowsAffected(); affected > 0 {
		created = true
	}
	var questionID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM questions WHERE composite_hash = ?`, compositeHash).Scan(&questionID); err != nil {
		return Question{}, false, err
	}
	if created {
		for position, option := range input.Options {
			if _, err := tx.ExecContext(ctx, `INSERT INTO question_options (question_id, option_key, option_text, position) VALUES (?, ?, ?, ?)`, questionID, strings.TrimSpace(option.Key), normalizeText(option.Text), position); err != nil {
				return Question{}, false, err
			}
		}
		answers := strings.Split(answerText, "###")
		raw := input.AnswerRaw
		if raw == "" {
			raw = input.Answer
		}
		for position, answer := range answers {
			if _, err := tx.ExecContext(ctx, `INSERT INTO question_answers (question_id, answer_text, answer_raw, position) VALUES (?, ?, ?, ?)`, questionID, answer, raw, position); err != nil {
				return Question{}, false, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return Question{}, false, err
	}
	question, err := s.GetByID(ctx, questionID)
	return question, created, err
}

func (s *Store) Search(ctx context.Context, query string) (Question, error) {
	normalized := normalizeText(query)
	if normalized == "" {
		return Question{}, ErrNotFound
	}
	var id uint64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM questions WHERE status = 1 AND question_hash = ? ORDER BY id ASC LIMIT 1`, hashText(normalized)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = s.db.QueryRowContext(ctx, `SELECT id FROM questions WHERE status = 1 AND INSTR(normalized_text, ?) > 0 ORDER BY CHAR_LENGTH(normalized_text), id ASC LIMIT 1`, normalized).Scan(&id)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return Question{}, ErrNotFound
	}
	if err != nil {
		return Question{}, err
	}
	return s.GetByID(ctx, id)
}

func (s *Store) GetByID(ctx context.Context, id uint64) (Question, error) {
	var question Question
	err := s.db.QueryRowContext(ctx, `SELECT id, question_text, question_type, platform, subject, source, collected_at FROM questions WHERE id = ?`, id).
		Scan(&question.ID, &question.Question, &question.Type, &question.Platform, &question.Subject, &question.Source, &question.CollectedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Question{}, ErrNotFound
	}
	if err != nil {
		return Question{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT option_key, option_text, position FROM question_options WHERE question_id = ? ORDER BY position ASC`, id)
	if err != nil {
		return Question{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.Key, &option.Text, &option.Position); err != nil {
			return Question{}, err
		}
		question.Options = append(question.Options, option)
	}
	answers, err := s.db.QueryContext(ctx, `SELECT answer_text, answer_raw, position FROM question_answers WHERE question_id = ? ORDER BY position ASC`, id)
	if err != nil {
		return Question{}, err
	}
	defer answers.Close()
	for answers.Next() {
		var answer Answer
		if err := answers.Scan(&answer.Text, &answer.Raw, &answer.Position); err != nil {
			return Question{}, err
		}
		question.Answers = append(question.Answers, answer)
	}
	return question, nil
}

func defaultType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "other"
	}
	return value
}
