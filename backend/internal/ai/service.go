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

type Service struct {
	store  *Store
	client *http.Client
}

func NewService(store *Store) *Service {
	return &Service{store: store, client: &http.Client{}}
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

func (s *Service) Solve(ctx context.Context, question, questionType string, options []string) (Answer, error) {
	hash := questionHash(question)
	if cached, err := s.store.GetCached(ctx, hash); err == nil {
		return cached, nil
	}
	models, err := s.store.ListModels(ctx)
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
	chargeCount := model.AIChargeCount
	if chargeCount <= 0 {
		chargeCount = 1
	}
	return Answer{QuestionHash: hash, Question: question, Type: questionType, Text: strings.TrimSpace(parsed.Choices[0].Message.Content), RawResponse: string(raw), Provider: model.ProviderName, Model: model.Name, TokenCount: parsed.Usage.TotalTokens, ChargeCount: chargeCount, Elapsed: time.Since(started)}, nil
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
