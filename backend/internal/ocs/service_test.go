package ocs

import "testing"

func TestUpdateSourceValidation(t *testing.T) {
	service := NewService(nil)
	if err := service.UpdateSourceStatus(nil, 1, 2); err == nil {
		t.Fatal("expected invalid source status to be rejected")
	}
	if err := service.UpdateSource(nil, 1, SourceInput{URL: "https://example.com"}); err == nil {
		t.Fatal("expected source name to be required")
	}
}

func TestLookupAndStringifyAnswer(t *testing.T) {
	document := map[string]any{
		"code":   float64(1),
		"data":   []any{"答案一", "答案二"},
		"nested": map[string]any{"items": []any{map[string]any{"value": "命中"}}},
	}
	if value, ok := lookup(document, "nested.items[0].value"); !ok || value != "命中" {
		t.Fatalf("unexpected nested lookup: %#v, %v", value, ok)
	}
	answer, err := stringifyAnswer(document["data"])
	if err != nil || answer != "答案一###答案二" {
		t.Fatalf("unexpected answer: %q, %v", answer, err)
	}
	if !equalValue(document["code"], "1") {
		t.Fatal("expected numeric success value to match")
	}
}

func TestReplacePlaceholders(t *testing.T) {
	got := replacePlaceholders("${title}|${question}|${type}|${options}", "题目", "single", "A. 一\nB. 二")
	if got != "题目|题目|single|A. 一\nB. 二" {
		t.Fatalf("unexpected replacement: %q", got)
	}
}

func TestResolveSafeCustomField(t *testing.T) {
	value, err := resolveFieldValue(map[string]any{
		"value":   "【单选题】${title}",
		"replace": []any{map[string]any{"from": "单选题", "to": ""}},
	}, "1 + 2 = ?", "single", "")
	if err != nil || value != "【】1 + 2 = ?" {
		t.Fatalf("unexpected replace result: %#v, %v", value, err)
	}
	mapped, err := resolveFieldValue(map[string]any{"value": "${type}", "map": map[string]any{"single": float64(1), "default": float64(0)}}, "", "single", "")
	if err != nil || mapped != float64(1) {
		t.Fatalf("unexpected map result: %#v, %v", mapped, err)
	}
	split, err := resolveFieldValue(map[string]any{"value": "${options}", "split": "\n"}, "", "", "A. 一\nB. 二")
	if err != nil || len(split.([]string)) != 2 {
		t.Fatalf("unexpected split result: %#v, %v", split, err)
	}
}

func TestRejectJavaScriptFieldHandler(t *testing.T) {
	if _, err := resolveFieldValue(map[string]any{"handler": "return (env) => env.title"}, "题目", "", ""); err == nil {
		t.Fatal("expected JavaScript handler to be rejected")
	}
}

func TestCanonicalAnswer(t *testing.T) {
	if got := canonicalAnswer("  Answer  Text "); got != "answer text" {
		t.Fatalf("unexpected canonical answer: %q", got)
	}
}
