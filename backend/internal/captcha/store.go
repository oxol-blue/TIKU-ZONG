package captcha

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"
)

const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

type item struct {
	code      string
	expiresAt time.Time
}

type Store struct {
	mu    sync.Mutex
	items map[string]item
}

func NewStore() *Store { return &Store{items: make(map[string]item)} }

func (s *Store) Generate() (id, image string, err error) {
	idBytes := make([]byte, 16)
	if _, err = rand.Read(idBytes); err != nil {
		return "", "", err
	}
	codeBytes := make([]byte, 6)
	if _, err = rand.Read(codeBytes); err != nil {
		return "", "", err
	}
	var code strings.Builder
	for _, value := range codeBytes {
		code.WriteByte(alphabet[int(value)%len(alphabet)])
	}
	id = base64.RawURLEncoding.EncodeToString(idBytes)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="160" height="48" viewBox="0 0 160 48"><rect width="160" height="48" rx="8" fill="#eef5ff"/><path d="M8 34 C35 8 60 45 88 16 S130 8 152 31" fill="none" stroke="#8ab4f8" stroke-width="2"/><text x="80" y="33" text-anchor="middle" font-family="Arial,sans-serif" font-size="23" font-weight="700" letter-spacing="4" fill="#2563eb">%s</text></svg>`, html.EscapeString(code.String()))
	s.mu.Lock()
	s.items[id] = item{code: strings.ToUpper(code.String()), expiresAt: time.Now().UTC().Add(5 * time.Minute)}
	s.mu.Unlock()
	return id, "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg)), nil
}

func (s *Store) Verify(id, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.items[id]
	delete(s.items, id)
	if !ok || time.Now().UTC().After(value.expiresAt) {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(code), value.code)
}
