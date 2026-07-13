package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/oxol-blue/TIKU-ZONG/backend/internal/captcha"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrInvalidInput       = errors.New("invalid input")
	ErrCaptchaRequired    = errors.New("captcha required or invalid")
	ErrSelfModification   = errors.New("administrator cannot disable or demote self")
)

type Service struct {
	store   *Store
	secret  string
	captcha *captcha.Store
}

func NewService(store *Store, secret string, captchaStores ...*captcha.Store) *Service {
	var captchaStore *captcha.Store
	if len(captchaStores) > 0 {
		captchaStore = captchaStores[0]
	}
	return &Service{store: store, secret: secret, captcha: captchaStore}
}

type Session struct {
	User         User   `json:"user"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

func (s *Service) Register(ctx context.Context, email, password string, inviteCodes ...string) (Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !validEmail(email) || len(password) < 8 || len(password) > 72 {
		return Session{}, ErrInvalidInput
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, err
	}
	inviteCode := ""
	if len(inviteCodes) > 0 {
		inviteCode = inviteCodes[0]
	}
	user, err := s.store.CreateUserWithInvite(ctx, email, string(passwordHash), inviteCode)
	if err != nil {
		return Session{}, err
	}
	return s.issueSession(ctx, user)
}

func (s *Service) CreateInvite(ctx context.Context, actorID uint64, input CreateInviteInput) (InviteView, error) {
	return s.store.CreateInvite(ctx, actorID, input)
}

func (s *Service) ListInvites(ctx context.Context) ([]InviteView, error) {
	return s.store.ListInvites(ctx)
}

func (s *Service) Login(ctx context.Context, email, password, captchaID, captchaCode string) (Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, passwordHash, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) {
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	if user.Status != 1 {
		return Session{}, ErrAccountDisabled
	}
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return Session{}, ErrAccountLocked
	}
	if user.FailedLoginCount >= 3 && (s.captcha == nil || !s.captcha.Verify(captchaID, captchaCode)) {
		return Session{}, ErrCaptchaRequired
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		var lockedUntil *time.Time
		if user.FailedLoginCount+1 >= 5 {
			until := time.Now().UTC().Add(15 * time.Minute)
			lockedUntil = &until
		}
		_ = s.store.RecordLoginFailure(ctx, user.ID, lockedUntil)
		if lockedUntil != nil {
			return Session{}, ErrAccountLocked
		}
		return Session{}, ErrInvalidCredentials
	}
	if err := s.store.RecordLoginSuccess(ctx, user.ID); err != nil {
		return Session{}, err
	}
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	return s.issueSession(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, plainRefreshToken string) (Session, error) {
	if plainRefreshToken == "" {
		return Session{}, ErrInvalidCredentials
	}
	userID, err := s.store.ConsumeRefreshToken(ctx, hashToken(plainRefreshToken))
	if err != nil {
		return Session{}, ErrInvalidCredentials
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil || user.Status != 1 {
		return Session{}, ErrInvalidCredentials
	}
	return s.issueSession(ctx, user)
}

func (s *Service) Authenticate(value string) (User, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return User{}, ErrInvalidCredentials
	}
	claims, err := parseAccessToken(s.secret, parts[1])
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	return s.store.GetUserByID(context.Background(), id)
}

func (s *Service) GetAPIKey(ctx context.Context, userID uint64) (APIKeyView, error) {
	return s.store.GetAPIKey(ctx, userID)
}

func (s *Service) CreateAPIKey(ctx context.Context, userID uint64) (string, APIKeyView, error) {
	return s.store.CreateAPIKey(ctx, userID)
}

func (s *Service) AuthenticateAPIKey(ctx context.Context, plain string) (User, uint64, error) {
	if strings.TrimSpace(plain) == "" {
		return User{}, 0, ErrInvalidCredentials
	}
	return s.store.ResolveAPIKey(ctx, plain)
}

func (s *Service) ListUsers(ctx context.Context, search string, status, page, pageSize int) (AdminUserPage, error) {
	return s.store.ListUsers(ctx, search, status, page, pageSize)
}

func (s *Service) UpdateUserStatus(ctx context.Context, actorID, userID uint64, status int) error {
	if status != 0 && status != 1 {
		return ErrInvalidInput
	}
	if actorID == userID && status == 0 {
		return ErrSelfModification
	}
	return s.store.UpdateStatus(ctx, userID, status)
}

func (s *Service) UpdateUserRole(ctx context.Context, actorID, userID uint64, role string) error {
	if role != RoleUser && role != RoleAdmin {
		return ErrInvalidInput
	}
	if actorID == userID && role != RoleAdmin {
		return ErrSelfModification
	}
	return s.store.UpdateRole(ctx, userID, role)
}

func (s *Service) issueSession(ctx context.Context, user User) (Session, error) {
	if s.secret == "" {
		return Session{}, fmt.Errorf("jwt secret is not configured")
	}
	accessToken, err := issueAccessToken(s.secret, user)
	if err != nil {
		return Session{}, err
	}
	refreshToken, refreshHash, err := newRefreshToken()
	if err != nil {
		return Session{}, err
	}
	if err := s.store.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().UTC().Add(refreshTokenTTL)); err != nil {
		return Session{}, err
	}
	return Session{User: user, AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int64(accessTokenTTL.Seconds())}, nil
}

func validEmail(value string) bool {
	return strings.Contains(value, "@") && !strings.HasPrefix(value, "@") && !strings.HasSuffix(value, "@")
}
