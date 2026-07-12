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
