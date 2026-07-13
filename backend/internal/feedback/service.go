package feedback

import (
	"context"
	"errors"
	"strings"
)

var validTypes = map[string]struct{}{
	"correct":     {},
	"incorrect":   {},
	"mismatch":    {},
	"parse_error": {},
	"other":       {},
}

type Service struct{ store *Store }

func NewService(store *Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID uint64, input Input) error {
	input.RequestID = strings.TrimSpace(input.RequestID)
	input.Question = strings.TrimSpace(input.Question)
	input.FeedbackType = strings.TrimSpace(input.FeedbackType)
	if input.RequestID == "" || input.Question == "" {
		return errors.New("requestId and question are required")
	}
	if _, ok := validTypes[input.FeedbackType]; !ok {
		return errors.New("invalid feedback type")
	}
	if len(input.Comment) > 1000 {
		return errors.New("comment is too long")
	}
	return s.store.Create(ctx, userID, input)
}
