package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("ai service unavailable")
var ErrQueueFull = errors.New("ai request queue is full")

type solveJob struct {
	ctx      context.Context
	question string
	typeName string
	options  []string
	result   chan solveResult
}

type solveResult struct {
	answer Answer
	err    error
}

type Service struct {
	store   *Store
	client  *http.Client
	jobs    chan solveJob
	workers int
}

type QueueStats struct {
	Depth    int `json:"depth"`
	Capacity int `json:"capacity"`
	Workers  int `json:"workers"`
}

func NewService(store *Store) *Service {
	return NewServiceWithQueue(store, 32, 2)
}

func NewServiceWithQueue(store *Store, queueSize, workers int) *Service {
	if queueSize <= 0 {
		queueSize = 1
	}
	if workers <= 0 {
		workers = 1
	}
	service := &Service{store: store, client: &http.Client{}, jobs: make(chan solveJob, queueSize), workers: workers}
	for index := 0; index < workers; index++ {
		go service.runWorker()
	}
	return service
}

func (s *Service) QueueStats() QueueStats {
	if s == nil || s.jobs == nil {
		return QueueStats{}
	}
	return QueueStats{Depth: len(s.jobs), Capacity: cap(s.jobs), Workers: s.workers}
}

func (s *Service) Cached(ctx context.Context, question string) (Answer, error) {
	return s.store.GetCached(ctx, questionHash(question))
}

func (s *Service) ListAnswers(ctx context.Context, search, provider, model string, status, page, pageSize int) (AnswerPage, error) {
	return s.store.ListAnswers(ctx, search, provider, model, status, page, pageSize)
}

func (s *Service) GetAnswer(ctx context.Context, id uint64) (AdminAnswer, error) {
	return s.store.GetAnswer(ctx, id)
}

func (s *Service) UpdateAnswerStatus(ctx context.Context, id uint64, status int) error {
	if status != 0 && status != 1 {
		return errors.New("invalid answer status")
	}
	return s.store.UpdateAnswerStatus(ctx, id, status)
}

func (s *Service) CreateModel(ctx context.Context, input CreateModelInput) (uint64, error) {
	if err := validateModelInput(input); err != nil {
		return 0, err
	}
	return s.store.CreateModel(ctx, input)
}

func (s *Service) UpdateModel(ctx context.Context, id uint64, input UpdateModelInput) error {
	if id == 0 {
		return ErrNotFound
	}
	if err := validateModelInput(CreateModelInput(input)); err != nil {
		return err
	}
	return s.store.UpdateModel(ctx, id, input)
}

func (s *Service) UpdateModelStatus(ctx context.Context, id uint64, status int) error {
	if id == 0 {
		return ErrNotFound
	}
	if status != 0 && status != 1 {
		return errors.New("invalid model status")
	}
	return s.store.UpdateModelStatus(ctx, id, status)
}

func validateModelInput(input CreateModelInput) error {
	if input.ProviderID == 0 || strings.TrimSpace(input.Name) == "" || input.AIChargeCount < 0 || input.TimeoutSeconds < 0 {
		return errors.New("invalid AI model")
	}
	mode := input.BillingMode
	if mode == "" {
		mode = BillingModeFixed
	}
	switch mode {
	case BillingModeFixed:
		return nil
	case BillingModeToken:
		if input.TokenUnit <= 0 {
			return errors.New("tokenUnit is required for token billing")
		}
	case BillingModeCost:
		if input.CostPerMillionTokensCents <= 0 || input.CostUnitCents <= 0 || input.CostMarkupPercent < 0 {
			return errors.New("cost configuration is required for cost billing")
		}
	default:
		return errors.New("invalid AI billing mode")
	}
	return nil
}

func chargeCount(model Model, tokenCount int) int {
	fixed := model.AIChargeCount
	if fixed <= 0 {
		fixed = 1
	}
	if tokenCount <= 0 {
		return fixed
	}
	switch model.BillingMode {
	case BillingModeToken:
		unit := model.TokenUnit
		if unit <= 0 {
			return fixed
		}
		return ceilDiv(tokenCount, unit) * fixed
	case BillingModeCost:
		if model.CostPerMillionTokensCents <= 0 || model.CostUnitCents <= 0 {
			return fixed
		}
		baseCost := ceilDiv64(int64(tokenCount)*int64(model.CostPerMillionTokensCents), 1_000_000)
		markup := ceilDiv64(baseCost*int64(model.CostMarkupPercent), 100)
		return int(maxInt64(1, ceilDiv64(baseCost+markup, int64(model.CostUnitCents))))
	default:
		return fixed
	}
}

func ceilDiv(value, divisor int) int { return int(ceilDiv64(int64(value), int64(divisor))) }

func ceilDiv64(value, divisor int64) int64 {
	if value <= 0 || divisor <= 0 {
		return 0
	}
	return 1 + (value-1)/divisor
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

func (s *Service) Solve(ctx context.Context, question, questionType string, options []string) (Answer, error) {
	if s.jobs == nil {
		return s.solveDirect(ctx, question, questionType, options)
	}
	result := make(chan solveResult, 1)
	job := solveJob{ctx: ctx, question: question, typeName: questionType, options: append([]string(nil), options...), result: result}
	select {
	case s.jobs <- job:
	case <-ctx.Done():
		return Answer{}, ctx.Err()
	default:
		return Answer{}, ErrQueueFull
	}
	select {
	case value := <-result:
		return value.answer, value.err
	case <-ctx.Done():
		return Answer{}, ctx.Err()
	}
}

func (s *Service) runWorker() {
	for job := range s.jobs {
		answer, err := s.solveDirect(job.ctx, job.question, job.typeName, job.options)
		job.result <- solveResult{answer: answer, err: err}
	}
}

func (s *Service) solveDirect(ctx context.Context, question, questionType string, options []string) (Answer, error) {
	hash := questionHash(question)
	if cached, err := s.store.GetCached(ctx, hash); err == nil {
		return cached, nil
	}
	models, err := s.store.ListActiveModels(ctx)
	if err != nil || len(models) == 0 {
		return Answer{}, ErrUnavailable
	}
	prompt := buildPrompt(question, questionType, options)
	for _, model := range models {
		answer, callErr := s.callModel(ctx, model, hash, question, questionType, prompt)
		if callErr == nil && strings.TrimSpace(answer.Text) != "" {
			if err := s.store.Save(ctx, answer, prompt); err != nil {
				return Answer{}, err
			}
			return answer, nil
		}
	}
	return Answer{}, ErrUnavailable
}

func (s *Service) callModel(parent context.Context, model Model, hash, question, questionType, prompt string) (Answer, error) {
	apiKey, err := s.store.DecryptKey(model.EncryptedKey)
	if err != nil {
		return Answer{}, err
	}
	timeout := time.Duration(model.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	body := map[string]any{
		"model":       model.Name,
		"temperature": 0,
		"messages":    []map[string]string{{"role": "system", "content": "You answer questions for a question bank. Return only the answer text. For multiple choice, return the option text, joining multiple answers with ###. Do not return option letters unless no option text is provided."}, {"role": "user", "content": prompt}},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Answer{}, err
	}
	url := strings.TrimRight(model.BaseURL, "/")
	if !strings.HasSuffix(url, "/chat/completions") {
		url += "/chat/completions"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Answer{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)
	started := time.Now()
	response, err := s.client.Do(request)
	if err != nil {
		return Answer{}, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return Answer{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Answer{}, fmt.Errorf("ai provider status %d", response.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return Answer{}, errors.New("invalid ai response")
	}
	return Answer{QuestionHash: hash, Question: question, Type: questionType, Text: strings.TrimSpace(parsed.Choices[0].Message.Content), RawResponse: string(raw), Provider: model.ProviderName, Model: model.Name, TokenCount: parsed.Usage.TotalTokens, ChargeCount: chargeCount(model, parsed.Usage.TotalTokens), Elapsed: time.Since(started)}, nil
}

func buildPrompt(question, questionType string, options []string) string {
	var builder strings.Builder
	builder.WriteString("Question type: ")
	builder.WriteString(questionType)
	builder.WriteString("\nQuestion: ")
	builder.WriteString(question)
	if len(options) > 0 {
		builder.WriteString("\nOptions:\n")
		builder.WriteString(strings.Join(options, "\n"))
	}
	return builder.String()
}

func questionHash(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
