package feedback

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var ErrDuplicate = errors.New("feedback already submitted")

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Create(ctx context.Context, userID uint64, input Input) error {
	digest := sha256.Sum256([]byte(strings.Join(strings.Fields(strings.TrimSpace(input.Question)), " ")))
	_, err := s.db.ExecContext(ctx, `INSERT INTO answer_feedbacks (user_id, request_id, question_hash, question_text, feedback_type, comment) VALUES (?, ?, ?, ?, ?, ?)`, userID, input.RequestID, hex.EncodeToString(digest[:]), input.Question, input.FeedbackType, input.Comment)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return ErrDuplicate
		}
		return fmt.Errorf("create answer feedback: %w", err)
	}
	return nil
}

func (s *Store) ListByUser(ctx context.Context, userID uint64, limit int) ([]Item, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, request_id, question_hash, feedback_type, comment, created_at FROM answer_feedbacks WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.RequestID, &item.QuestionHash, &item.FeedbackType, &item.Comment, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListAdmin(ctx context.Context, feedbackType, search string, page, pageSize int) (Page, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	where := "WHERE 1 = 1"
	args := make([]any, 0, 3)
	if strings.TrimSpace(feedbackType) != "" {
		where += " AND f.feedback_type = ?"
		args = append(args, strings.TrimSpace(feedbackType))
	}
	if strings.TrimSpace(search) != "" {
		where += " AND (f.request_id LIKE ? OR f.question_hash LIKE ? OR u.email LIKE ?)"
		value := "%" + strings.TrimSpace(search) + "%"
		args = append(args, value, value, value)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM answer_feedbacks f JOIN users u ON u.id = f.user_id "+where, args...).Scan(&total); err != nil {
		return Page{}, err
	}
	queryArgs := append(append([]any{}, args...), (page-1)*pageSize, pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT f.id, f.request_id, f.question_hash, f.feedback_type, f.comment, f.created_at, f.user_id, u.email
		FROM answer_feedbacks f JOIN users u ON u.id = f.user_id `+where+` ORDER BY f.id DESC LIMIT ?, ?`, queryArgs...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	items := make([]AdminItem, 0)
	for rows.Next() {
		var item AdminItem
		if err := rows.Scan(&item.ID, &item.RequestID, &item.QuestionHash, &item.FeedbackType, &item.Comment, &item.CreatedAt, &item.UserID, &item.UserEmail); err != nil {
			return Page{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	return Page{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}
