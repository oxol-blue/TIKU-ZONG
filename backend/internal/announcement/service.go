package announcement

import (
	"context"
	"errors"
	"strings"
)

type Service struct{ store *Store }

func NewService(store *Store) *Service { return &Service{store: store} }

func (s *Service) ListPublished(ctx context.Context) ([]Item, error) {
	return s.store.ListPublished(ctx)
}
func (s *Service) ListAll(ctx context.Context) ([]Item, error) { return s.store.ListAll(ctx) }

func (s *Service) Create(ctx context.Context, input CreateInput) (Item, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Title == "" || input.Content == "" || input.IsPinned < 0 || input.IsPinned > 1 {
		return Item{}, errors.New("invalid announcement")
	}
	return s.store.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uint64, input UpdateInput) (Item, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Title == "" || input.Content == "" || input.IsPinned < 0 || input.IsPinned > 1 {
		return Item{}, errors.New("invalid announcement")
	}
	return s.store.Update(ctx, id, input)
}

func (s *Service) UpdateStatus(ctx context.Context, id uint64, status int) error {
	if status != 0 && status != 1 {
		return errors.New("invalid announcement status")
	}
	return s.store.UpdateStatus(ctx, id, status)
}
