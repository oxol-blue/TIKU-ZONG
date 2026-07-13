package calls

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type Log struct {
	RequestID  string
	UserID     uint64
	APIKeyID   uint64
	Endpoint   string
	Question   string
	Success    bool
	IsAI       bool
	SourceKind string
	Elapsed    time.Duration
	HTTPStatus int
	ErrorCode  string
}

// SearchHistory is a successful search result kept for the requesting user.
// Unlike api_call_logs, it retains readable question and answer text so the user
// can revisit their own online and API searches.
type SearchHistory struct {
	UserID    uint64
	RequestID string
	Question  string
	Type      string
	Answer    string
	Source    string
	IsAI      bool
	Elapsed   time.Duration
}

type SearchHistoryItem struct {
	ID            uint64    `json:"id"`
	RequestID     string    `json:"requestId"`
	Question      string    `json:"question"`
	Type          string    `json:"type"`
	Answer        string    `json:"answer"`
	Source        string    `json:"source"`
	IsAI          bool      `json:"isAi"`
	ElapsedMicros int64     `json:"elapsedMicros"`
	CreatedAt     time.Time `json:"createdAt"`
}

type SearchHistoryPage struct {
	Items    []SearchHistoryItem `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int                 `json:"total"`
}

type Store struct{ db *sql.DB }

type Dashboard struct {
	UserCount           int64   `json:"userCount"`
	PaidUserCount       int64   `json:"paidUserCount"`
	PaidOrderCount      int64   `json:"paidOrderCount"`
	PaidAmountCents     int64   `json:"paidAmountCents"`
	CallCount           int64   `json:"callCount"`
	SuccessfulCalls     int64   `json:"successfulCalls"`
	AICallCount         int64   `json:"aiCallCount"`
	OCSCallCount        int64   `json:"ocsCallCount"`
	OnlineSearchCount   int64   `json:"onlineSearchCount"`
	LocalHitCount       int64   `json:"localHitCount"`
	OCSHitCount         int64   `json:"ocsHitCount"`
	TokenCount          int64   `json:"tokenCount"`
	PackageConsumeCount int64   `json:"packageConsumeCount"`
	ErrorRate           float64 `json:"errorRate"`
	AverageLatencyMs    float64 `json:"averageLatencyMs"`
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Log(ctx context.Context, item Log) error {
	digest := sha256.Sum256([]byte(item.Question))
	var userID any
	if item.UserID != 0 {
		userID = item.UserID
	}
	var keyID any
	if item.APIKeyID != 0 {
		keyID = item.APIKeyID
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO api_call_logs (request_id, user_id, api_key_id, endpoint, question_hash, success, is_ai, source_kind, elapsed_micros, http_status, error_code) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.RequestID, userID, keyID, item.Endpoint, hex.EncodeToString(digest[:]), boolInt(item.Success), boolInt(item.IsAI), item.SourceKind, item.Elapsed.Microseconds(), item.HTTPStatus, item.ErrorCode)
	if err != nil {
		return fmt.Errorf("write api call log: %w", err)
	}
	return nil
}

func (s *Store) RecordSearch(ctx context.Context, item SearchHistory) error {
	if s == nil || s.db == nil || item.UserID == 0 || item.RequestID == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO search_history (user_id, request_id, question_text, question_type, answer_text, source, is_ai, elapsed_micros) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id`,
		item.UserID, item.RequestID, item.Question, item.Type, item.Answer, item.Source, boolInt(item.IsAI), item.Elapsed.Microseconds())
	if err != nil {
		return fmt.Errorf("write search history: %w", err)
	}
	return nil
}

func (s *Store) SearchHistoryByUser(ctx context.Context, userID uint64, isAI *bool, page, pageSize int) (SearchHistoryPage, error) {
	page, pageSize = normalizeHistoryPage(page, pageSize)
	where := "WHERE user_id = ?"
	args := []any{userID}
	if isAI != nil {
		where += " AND is_ai = ?"
		args = append(args, boolInt(*isAI))
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM search_history "+where, args...).Scan(&total); err != nil {
		return SearchHistoryPage{}, err
	}
	args = append(args, (page-1)*pageSize, pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT id, request_id, question_text, question_type, answer_text, source, is_ai, elapsed_micros, created_at FROM search_history `+where+` ORDER BY id DESC LIMIT ?, ?`, args...)
	if err != nil {
		return SearchHistoryPage{}, err
	}
	defer rows.Close()
	items := make([]SearchHistoryItem, 0)
	for rows.Next() {
		var item SearchHistoryItem
		var isAI int
		if err := rows.Scan(&item.ID, &item.RequestID, &item.Question, &item.Type, &item.Answer, &item.Source, &isAI, &item.ElapsedMicros, &item.CreatedAt); err != nil {
			return SearchHistoryPage{}, err
		}
		item.IsAI = isAI == 1
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SearchHistoryPage{}, err
	}
	return SearchHistoryPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Store) Recent(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT request_id, user_id, api_key_id, endpoint, question_hash, success, is_ai, source_kind, elapsed_micros, http_status, error_code, created_at FROM api_call_logs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var requestID, endpoint, questionHash, sourceKind, errorCode string
		var userID, keyID, elapsed sql.NullInt64
		var success, isAI, status int
		var createdAt time.Time
		if err := rows.Scan(&requestID, &userID, &keyID, &endpoint, &questionHash, &success, &isAI, &sourceKind, &elapsed, &status, &errorCode, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"requestId": requestID, "userId": nullableInt(userID), "apiKeyId": nullableInt(keyID), "endpoint": endpoint, "questionHash": questionHash, "success": success == 1, "isAi": isAI == 1, "sourceKind": sourceKind, "elapsedMicros": elapsed.Int64, "httpStatus": status, "errorCode": errorCode, "createdAt": createdAt})
	}
	return items, rows.Err()
}

func (s *Store) RecentByUser(ctx context.Context, userID uint64, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT request_id, endpoint, question_hash, success, is_ai, source_kind, elapsed_micros, http_status, error_code, created_at FROM api_call_logs WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var requestID, endpoint, questionHash, sourceKind, errorCode string
		var elapsed sql.NullInt64
		var success, isAI, status int
		var createdAt time.Time
		if err := rows.Scan(&requestID, &endpoint, &questionHash, &success, &isAI, &sourceKind, &elapsed, &status, &errorCode, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"requestId": requestID, "endpoint": endpoint, "questionHash": questionHash, "success": success == 1, "isAi": isAI == 1, "sourceKind": sourceKind, "elapsedMicros": elapsed.Int64, "httpStatus": status, "errorCode": errorCode, "createdAt": createdAt})
	}
	return items, rows.Err()
}

func (s *Store) Dashboard(ctx context.Context) (Dashboard, error) {
	var result Dashboard
	queries := []struct {
		destination any
		query       string
	}{
		{&result.UserCount, `SELECT COUNT(*) FROM users`},
		{&result.PaidUserCount, `SELECT COUNT(DISTINCT user_id) FROM payment_orders WHERE status IN ('paid', 'partial_refunded', 'refunded')`},
		{&result.PaidOrderCount, `SELECT COUNT(*) FROM payment_orders WHERE status IN ('paid', 'partial_refunded', 'refunded')`},
		{&result.PaidAmountCents, `SELECT COALESCE(SUM(payable_cents), 0) FROM payment_orders WHERE status IN ('paid', 'partial_refunded', 'refunded')`},
		{&result.CallCount, `SELECT COUNT(*) FROM api_call_logs`},
		{&result.SuccessfulCalls, `SELECT COUNT(*) FROM api_call_logs WHERE success = 1`},
		{&result.AICallCount, `SELECT COUNT(*) FROM api_call_logs WHERE is_ai = 1`},
		{&result.OCSCallCount, `SELECT COUNT(*) FROM api_call_logs WHERE endpoint = '/api/ocs/search'`},
		{&result.OnlineSearchCount, `SELECT COUNT(*) FROM api_call_logs WHERE endpoint = '/api/v1/search' AND api_key_id IS NULL`},
		{&result.LocalHitCount, `SELECT COUNT(*) FROM api_call_logs WHERE success = 1 AND source_kind = 'local'`},
		{&result.OCSHitCount, `SELECT COUNT(*) FROM api_call_logs WHERE success = 1 AND source_kind = 'ocs'`},
		{&result.TokenCount, `SELECT COALESCE(SUM(token_count), 0) FROM question_ai`},
		{&result.PackageConsumeCount, `SELECT COALESCE(SUM(amount), 0) FROM package_consumptions`},
	}
	for _, item := range queries {
		if err := s.db.QueryRowContext(ctx, item.query).Scan(item.destination); err != nil {
			return Dashboard{}, err
		}
	}
	var averageMicros sql.NullFloat64
	if err := s.db.QueryRowContext(ctx, `SELECT AVG(elapsed_micros) FROM api_call_logs`).Scan(&averageMicros); err != nil {
		return Dashboard{}, err
	}
	if averageMicros.Valid {
		result.AverageLatencyMs = averageMicros.Float64 / 1000
	}
	if result.CallCount > 0 {
		result.ErrorRate = float64(result.CallCount-result.SuccessfulCalls) * 100 / float64(result.CallCount)
	}
	return result, nil
}

func nullableInt(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func normalizeHistoryPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
