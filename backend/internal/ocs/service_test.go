package ocs

import "testing"

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

func TestCanonicalAnswer(t *testing.T) {
	if got := canonicalAnswer("  Answer  Text "); got != "answer text" {
		t.Fatalf("unexpected canonical answer: %q", got)
	}
}
