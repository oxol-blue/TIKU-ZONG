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
