package questions

import (
	"context"
	"testing"
)

func TestAdminQuestionStatusValidation(t *testing.T) {
	service := &Service{}
	if err := service.UpdateStatus(context.Background(), 1, 2); err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestQuestionAdminPageLimits(t *testing.T) {
	if page, size := normalizePage(0, 0); page != 1 || size != 20 {
		t.Fatalf("unexpected defaults: page=%d size=%d", page, size)
	}
	if page, size := normalizePage(3, 1000); page != 3 || size != 100 {
		t.Fatalf("unexpected limits: page=%d size=%d", page, size)
	}
}

func TestSimilarityScore(t *testing.T) {
	if score := similarityScore("这是一个测试题目", "这是一个测试题目"); score != 1 {
		t.Fatalf("expected exact score 1, got %f", score)
	}
	if score := similarityScore("这是一个测试题目", "完全不同的问题"); score >= 0.35 {
		t.Fatalf("expected unrelated score below threshold, got %f", score)
	}
	if score := similarityScore("下列哪项属于测试", "下列哪项属于测试内容"); score < 0.35 {
		t.Fatalf("expected related score above threshold, got %f", score)
	}
}
