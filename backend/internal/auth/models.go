package auth

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID               uint64     `json:"id"`
	Email            string     `json:"email"`
	Role             string     `json:"role"`
	Status           int        `json:"status"`
	FailedLoginCount int        `json:"-"`
	LockedUntil      *time.Time `json:"-"`
	LastLoginAt      *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type APIKeyView struct {
	Prefix     string     `json:"prefix"`
	Masked     string     `json:"masked"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}
