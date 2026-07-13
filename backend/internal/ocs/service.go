package ocs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrNoAnswer = errors.New("ocs source returned no answer")

type Service struct {
	store  *Store
	client *http.Client
}

func NewService(store *Store) *Service {
	return &Service{store: store, client: &http.Client{Timeout: 6 * time.Second}}
}

func (s *Service) CreateSource(ctx context.Context, input SourceInput) (Source, error) {
	if err := validateURL(input.URL); err != nil {
		return Source{}, err
	}
	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))
	if input.Method == "" {
		input.Method = http.MethodGet
	}
	if input.Method != http.MethodGet && input.Method != http.MethodPost {
		return Source{}, errors.New("OCS method must be GET or POST")
	}
	if input.Priority <= 0 {
		input.Priority = 100
	}
	return s.store.CreateSource(ctx, input)
}

func (s *Service) ListSources(ctx context.Context) ([]Source, error) { return s.store.ListSources(ctx) }

func (s *Service) Search(ctx context.Context, question, questionType string, options []string) (Result, error) {
	sources, err := s.store.ListEnabledSources(ctx)
	if err != nil {
		return Result{}, err
	}
	for _, source := range sources {
		result, callErr := s.call(ctx, source, question, questionType, options)
		if callErr == nil && strings.TrimSpace(result.Answer) != "" {
			return result, nil
		}
	}
	return Result{}, ErrNoAnswer
}

func (s *Service) call(parent context.Context, source Source, question, questionType string, options []string) (Result, error) {
	started := time.Now()
	values := make(map[string]string, len(source.Data))
	for key, value := range source.Data {
		values[key] = replacePlaceholders(value, question, questionType, strings.Join(options, "\n"))
	}
	endpoint := source.URL
	var body io.Reader
	method := strings.ToUpper(source.Method)
	if method == http.MethodGet {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return Result{}, err
		}
		query := parsed.Query()
		for key, value := range values {
			query.Set(key, value)
		}
		parsed.RawQuery = query.Encode()
		endpoint = parsed.String()
	} else {
		payload, err := json.Marshal(values)
		if err != nil {
			return Result{}, err
		}
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(parent, method, endpoint, body)
	if err != nil {
		return Result{}, err
	}
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range source.Headers {
		request.Header.Set(key, replacePlaceholders(value, question, questionType, strings.Join(options, "\n")))
	}
	response, err := s.client.Do(request)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{}, fmt.Errorf("OCS source returned HTTP %d", response.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return Result{}, err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return Result{}, err
	}
	if source.SuccessPath != "" {
		value, ok := lookup(document, source.SuccessPath)
		if !ok || (source.SuccessValue != "" && !equalValue(value, source.SuccessValue)) {
			return Result{}, ErrNoAnswer
		}
	}
	answerValue, ok := lookup(document, source.AnswerPath)
	if !ok {
		return Result{}, ErrNoAnswer
	}
	answer, err := stringifyAnswer(answerValue)
	if err != nil || strings.TrimSpace(answer) == "" {
		return Result{}, ErrNoAnswer
	}
	returnedQuestion := question
	if value, found := lookup(document, source.QuestionPath); found {
		if text, textErr := stringifyAnswer(value); textErr == nil && strings.TrimSpace(text) != "" {
			returnedQuestion = text
		}
	}
	return Result{Question: returnedQuestion, Answer: answer, Source: source.Name, Elapsed: time.Since(started)}, nil
}

func validateURL(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return errors.New("OCS URL must be an absolute HTTP or HTTPS URL")
	}
	return nil
}

func replacePlaceholders(value, question, questionType, options string) string {
	replacer := strings.NewReplacer("${title}", question, "${question}", question, "${type}", questionType, "${options}", options)
	return replacer.Replace(value)
}

func lookup(document any, path string) (any, bool) {
	if strings.TrimSpace(path) == "" {
		return document, true
	}
	current := document
	for _, part := range strings.Split(path, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if indexStart := strings.Index(part, "["); indexStart >= 0 && strings.HasSuffix(part, "]") {
			name := part[:indexStart]
			index, err := strconv.Atoi(part[indexStart+1 : len(part)-1])
			if err != nil {
				return nil, false
			}
			if name != "" {
				var ok bool
				current, ok = lookup(current, name)
				if !ok {
					return nil, false
				}
			}
			items, ok := current.([]any)
			if !ok || index < 0 || index >= len(items) {
				return nil, false
			}
			current = items[index]
			continue
		}
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func equalValue(value any, expected string) bool {
	return strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), strings.TrimSpace(expected))
}

func stringifyAnswer(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			text, err := stringifyAnswer(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, text)
		}
		return strings.Join(parts, "###"), nil
	case float64, bool:
		return fmt.Sprint(typed), nil
	default:
		encoded, err := json.Marshal(value)
		return string(encoded), err
	}
}
