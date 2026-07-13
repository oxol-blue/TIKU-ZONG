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
	store     *Store
	client    *http.Client
	mergeRule string
}

func NewService(store *Store, mergeRules ...string) *Service {
	rule := "priority"
	if len(mergeRules) > 0 && strings.EqualFold(strings.TrimSpace(mergeRules[0]), "majority") {
		rule = "majority"
	}
	return &Service{store: store, client: &http.Client{Timeout: 6 * time.Second}, mergeRule: rule}
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
	if err := validateFieldDefinitions(input.Data); err != nil {
		return Source{}, err
	}
	return s.store.CreateSource(ctx, input)
}

func (s *Service) ListSources(ctx context.Context) ([]Source, error) { return s.store.ListSources(ctx) }

func (s *Service) Search(ctx context.Context, question, questionType string, options []string) (Result, error) {
	sources, err := s.store.ListEnabledSources(ctx)
	if err != nil {
		return Result{}, err
	}
	if s.mergeRule != "majority" {
		for _, source := range sources {
			result, callErr := s.call(ctx, source, question, questionType, options)
			if callErr == nil && strings.TrimSpace(result.Answer) != "" {
				return result, nil
			}
		}
		return Result{}, ErrNoAnswer
	}
	return s.majoritySearch(ctx, sources, question, questionType, options)
}

func (s *Service) majoritySearch(ctx context.Context, sources []Source, question, questionType string, options []string) (Result, error) {
	type candidate struct {
		result Result
		count  int
		order  int
	}
	candidates := make([]candidate, 0, len(sources))
	counts := make(map[string]int, len(sources))
	for order, source := range sources {
		result, callErr := s.call(ctx, source, question, questionType, options)
		if callErr == nil && strings.TrimSpace(result.Answer) != "" {
			key := canonicalAnswer(result.Answer)
			counts[key]++
			candidates = append(candidates, candidate{result: result, order: order})
		}
	}
	best := -1
	for index := range candidates {
		candidates[index].count = counts[canonicalAnswer(candidates[index].result.Answer)]
		if best < 0 || candidates[index].count > candidates[best].count || (candidates[index].count == candidates[best].count && candidates[index].order < candidates[best].order) {
			best = index
		}
	}
	if best < 0 {
		return Result{}, ErrNoAnswer
	}
	return candidates[best].result, nil
}

func canonicalAnswer(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func (s *Service) call(parent context.Context, source Source, question, questionType string, options []string) (Result, error) {
	started := time.Now()
	values := make(map[string]any, len(source.Data))
	for key, value := range source.Data {
		resolved, err := resolveFieldValue(value, question, questionType, strings.Join(options, "\n"))
		if err != nil {
			return Result{}, err
		}
		values[key] = resolved
	}
	endpoint := replacePlaceholders(source.URL, question, questionType, strings.Join(options, "\n"))
	var body io.Reader
	method := strings.ToUpper(source.Method)
	if method == http.MethodGet {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return Result{}, err
		}
		query := parsed.Query()
		for key, value := range values {
			encoded, encodeErr := encodeQueryValue(value)
			if encodeErr != nil {
				return Result{}, encodeErr
			}
			query.Set(key, encoded)
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

// resolveFieldValue implements the safe subset of OCS custom-field handlers.
// It deliberately does not evaluate JavaScript. A field may be a plain JSON
// value or a DSL object with value/template plus replace, map, split and join.
func resolveFieldValue(value any, question, questionType, options string) (any, error) {
	switch typed := value.(type) {
	case string:
		return replacePlaceholders(typed, question, questionType, options), nil
	case map[string]any:
		if _, exists := typed["handler"]; exists {
			return nil, errors.New("OCS JavaScript field handlers are not supported; use the safe value/replace/map/split/join DSL")
		}
		raw, exists := typed["value"]
		if !exists {
			raw, exists = typed["template"]
		}
		if !exists {
			return nil, errors.New("OCS custom field requires value or template")
		}
		resolved, err := resolveFieldValue(raw, question, questionType, options)
		if err != nil {
			return nil, err
		}
		text, isText := resolved.(string)
		if replacements, found := typed["replace"]; found {
			if !isText {
				return nil, errors.New("OCS replace only supports string values")
			}
			text, err = applyReplacements(text, replacements)
			if err != nil {
				return nil, err
			}
			resolved, isText = text, true
		}
		if mapping, found := typed["map"]; found {
			if !isText {
				return nil, errors.New("OCS map only supports string values")
			}
			mapped, mapErr := applyMapping(text, mapping)
			if mapErr != nil {
				return nil, mapErr
			}
			resolved = mapped
			text, isText = mapped.(string)
		}
		if delimiter, found := typed["split"]; found {
			if !isText {
				return nil, errors.New("OCS split only supports string values")
			}
			delimiterText, ok := delimiter.(string)
			if !ok {
				return nil, errors.New("OCS split must be a string")
			}
			resolved = strings.Split(text, delimiterText)
		}
		if join, found := typed["join"]; found {
			delimiter, ok := join.(string)
			if !ok {
				return nil, errors.New("OCS join must be a string")
			}
			items, ok := resolved.([]string)
			if !ok {
				return nil, errors.New("OCS join requires a split result")
			}
			resolved = strings.Join(items, delimiter)
		}
		return resolved, nil
	default:
		return value, nil
	}
}

func applyReplacements(value string, definition any) (string, error) {
	items, ok := definition.([]any)
	if !ok {
		return "", errors.New("OCS replace must be an array")
	}
	for _, item := range items {
		rule, ok := item.(map[string]any)
		if !ok {
			return "", errors.New("OCS replace items must be objects")
		}
		from, fromOK := rule["from"].(string)
		to, toOK := rule["to"].(string)
		if !fromOK || !toOK {
			return "", errors.New("OCS replace items require string from and to")
		}
		value = strings.ReplaceAll(value, from, to)
	}
	return value, nil
}

func applyMapping(value string, definition any) (any, error) {
	mapping, ok := definition.(map[string]any)
	if !ok {
		return nil, errors.New("OCS map must be an object")
	}
	if mapped, found := mapping[value]; found {
		return mapped, nil
	}
	if fallback, found := mapping["default"]; found {
		return fallback, nil
	}
	return value, nil
}

func encodeQueryValue(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case float64, bool, int, int64:
		return fmt.Sprint(typed), nil
	default:
		encoded, err := json.Marshal(value)
		return string(encoded), err
	}
}

func validateFieldDefinitions(values map[string]any) error {
	for _, value := range values {
		if _, err := resolveFieldValue(value, "", "", ""); err != nil {
			return err
		}
	}
	return nil
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
