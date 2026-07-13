package calls

import "testing"

func TestNormalizeHistoryPage(t *testing.T) {
	tests := []struct {
		page, pageSize     int
		wantPage, wantSize int
	}{
		{0, 0, 1, 20},
		{3, 10, 3, 10},
		{1, 1000, 1, 100},
	}
	for _, test := range tests {
		page, pageSize := normalizeHistoryPage(test.page, test.pageSize)
		if page != test.wantPage || pageSize != test.wantSize {
			t.Fatalf("normalizeHistoryPage(%d, %d) = (%d, %d), want (%d, %d)", test.page, test.pageSize, page, pageSize, test.wantPage, test.wantSize)
		}
	}
}
