package questions

import "testing"

func TestNormalizeAnswerMapsOptionKeysToText(t *testing.T) {
	options := []OptionInput{{Key: "A", Text: "北京"}, {Key: "C", Text: "广州"}}
	if got := normalizeAnswer("A###C", options); got != "北京###广州" {
		t.Fatalf("expected option text answer, got %q", got)
	}
}

func TestNormalizeAnswerSortsMultipleAnswers(t *testing.T) {
	options := []OptionInput{{Key: "A", Text: "北京"}, {Key: "C", Text: "广州"}}
	if got := normalizeAnswer("C###A", options); got != "北京###广州" {
		t.Fatalf("expected stable multiple-answer order, got %q", got)
	}
}

func TestNormalizeTextCollapsesWhitespace(t *testing.T) {
	if got := normalizeText("  hello\n\tworld  "); got != "hello world" {
		t.Fatalf("expected normalized whitespace, got %q", got)
	}
}
