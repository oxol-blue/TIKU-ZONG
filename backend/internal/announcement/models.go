package announcement

import "time"

type Item struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Status      int        `json:"status"`
	IsPinned    int        `json:"isPinned"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type CreateInput struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsPinned int    `json:"isPinned"`
	Status   int    `json:"status"`
}

type UpdateInput struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsPinned int    `json:"isPinned"`
}
