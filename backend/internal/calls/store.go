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
	Elapsed    time.Duration
	HTTPStatus int
	ErrorCode  string
}

type Store struct{ db *sql.DB }

type Dashboard struct {
	UserCount        int64   `json:"userCount"`
	PaidUserCount    int64   `json:"paidUserCount"`
	PaidOrderCount   int64   `json:"paidOrderCount"`
	PaidAmountCents  int64   `json:"paidAmountCents"`
	CallCount        int64   `json:"callCount"`
	SuccessfulCalls  int64   `json:"successfulCalls"`
	AICallCount      int64   `json:"aiCallCount"`
	OCSCallCount     int64   `json:"ocsCallCount"`
	AverageLatencyMs float64 `json:"averageLatencyMs"`
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
	_, err := s.db.ExecContext(ctx, `INSERT INTO api_call_logs (request_id, user_id, api_key_id, endpoint, question_hash, success, is_ai, elapsed_micros, http_status, error_code) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.RequestID, userID, keyID, item.Endpoint, hex.EncodeToString(digest[:]), boolInt(item.Success), boolInt(item.IsAI), item.Elapsed.Microseconds(), item.HTTPStatus, item.ErrorCode)
	if err != nil {
		return fmt.Errorf("write api call log: %w", err)
	}
	return nil
}

func (s *Store) Recent(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT request_id, user_id, api_key_id, endpoint, question_hash, success, is_ai, elapsed_micros, http_status, error_code, created_at FROM api_call_logs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var requestID, endpoint, questionHash, errorCode string
		var userID, keyID, elapsed sql.NullInt64
		var success, isAI, status int
		var createdAt time.Time
		if err := rows.Scan(&requestID, &userID, &keyID, &endpoint, &questionHash, &success, &isAI, &elapsed, &status, &errorCode, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"requestId": requestID, "userId": nullableInt(userID), "apiKeyId": nullableInt(keyID), "endpoint": endpoint, "questionHash": questionHash, "success": success == 1, "isAi": isAI == 1, "elapsedMicros": elapsed.Int64, "httpStatus": status, "errorCode": errorCode, "createdAt": createdAt})
	}
	return items, rows.Err()
}

func (s *Store) RecentByUser(ctx context.Context, userID uint64, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT request_id, endpoint, question_hash, success, is_ai, elapsed_micros, http_status, error_code, created_at FROM api_call_logs WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var requestID, endpoint, questionHash, errorCode string
		var elapsed sql.NullInt64
		var success, isAI, status int
		var createdAt time.Time
		if err := rows.Scan(&requestID, &endpoint, &questionHash, &success, &isAI, &elapsed, &status, &errorCode, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"requestId": requestID, "endpoint": endpoint, "questionHash": questionHash, "success": success == 1, "isAi": isAI == 1, "elapsedMicros": elapsed.Int64, "httpStatus": status, "errorCode": errorCode, "createdAt": createdAt})
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
