package announcement

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("announcement not found")

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func scan(row interface{ Scan(...any) error }) (Item, error) {
	var item Item
	err := row.Scan(&item.ID, &item.Title, &item.Content, &item.Status, &item.IsPinned, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	return item, err
}

const selectColumns = `id, title, content, status, is_pinned, published_at, created_at, updated_at`

func (s *Store) Get(ctx context.Context, id uint64) (Item, error) {
	return scan(s.db.QueryRowContext(ctx, `SELECT `+selectColumns+` FROM announcements WHERE id = ?`, id))
}

func (s *Store) ListPublished(ctx context.Context) ([]Item, error) {
	return s.list(ctx, `WHERE status = 1 AND (published_at IS NULL OR published_at <= UTC_TIMESTAMP(6)) ORDER BY is_pinned DESC, published_at DESC, id DESC`)
}

func (s *Store) ListAll(ctx context.Context) ([]Item, error) {
	return s.list(ctx, `ORDER BY is_pinned DESC, id DESC`)
}

func (s *Store) list(ctx context.Context, suffix string) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+selectColumns+` FROM announcements `+suffix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) Create(ctx context.Context, input CreateInput) (Item, error) {
	status := input.Status
	if status != 0 && status != 1 {
		status = 1
	}
	var publishedAt *time.Time
	if status == 1 {
		now := time.Now().UTC()
		publishedAt = &now
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO announcements (title, content, status, is_pinned, published_at) VALUES (?, ?, ?, ?, ?)`, input.Title, input.Content, status, input.IsPinned, publishedAt)
	if err != nil {
		return Item{}, fmt.Errorf("create announcement: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Item{}, err
	}
	return s.Get(ctx, uint64(id))
}

func (s *Store) Update(ctx context.Context, id uint64, input UpdateInput) (Item, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE announcements SET title = ?, content = ?, is_pinned = ? WHERE id = ?`, input.Title, input.Content, input.IsPinned, id)
	if err != nil {
		return Item{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Item{}, ErrNotFound
	}
	return s.Get(ctx, id)
}

func (s *Store) UpdateStatus(ctx context.Context, id uint64, status int) error {
	var publishedAt any
	if status == 1 {
		publishedAt = time.Now().UTC()
	} else {
		publishedAt = nil
	}
	result, err := s.db.ExecContext(ctx, `UPDATE announcements SET status = ?, published_at = ? WHERE id = ?`, status, publishedAt, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}
