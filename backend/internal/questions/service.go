package questions

import (
	"context"
	"errors"
	"time"
)

type Service struct{ store *Store }

func NewService(store *Store) *Service { return &Service{store: store} }

func (s *Service) Import(ctx context.Context, items []ImportInput) (created, duplicates int, err error) {
	for _, item := range items {
		if item.Question == "" || item.Answer == "" {
			return created, duplicates, errors.New("question and answer are required")
		}
		_, wasCreated, err := s.store.Upsert(ctx, item)
		if err != nil {
			return created, duplicates, err
		}
		if wasCreated {
			created++
		} else {
			duplicates++
		}
	}
	return created, duplicates, nil
}

func (s *Service) Search(ctx context.Context, query string) (Question, time.Duration, error) {
	started := time.Now()
	question, err := s.store.Search(ctx, query)
	return question, time.Since(started), err
}

func (s *Service) ListAdmin(ctx context.Context, search, questionType, subject string, status, page, pageSize int) (QuestionPage, error) {
	return s.store.ListAdmin(ctx, search, questionType, subject, status, page, pageSize)
}

func (s *Service) GetByID(ctx context.Context, id uint64) (Question, error) {
	return s.store.GetByID(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id uint64, status int) error {
	if status != 0 && status != 1 {
		return errors.New("invalid question status")
	}
	return s.store.UpdateStatus(ctx, id, status)
}
