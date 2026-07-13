package ai

import "testing"

func TestBuildPrompt(t *testing.T) {
	prompt := buildPrompt("题目", "single", []string{"A. 北京", "B. 上海"})
	if prompt == "" || prompt == "题目" {
		t.Fatalf("expected prompt context, got %q", prompt)
	}
}
