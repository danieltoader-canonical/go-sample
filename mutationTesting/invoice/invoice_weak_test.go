//go:build weak

package invoice

import "testing"

func TestIsOverdue(t *testing.T) {
	tests := []struct {
		balance     int64
		daysPastDue int
		want        bool
	}{
		{100, 1, true},
		{100, -1, false},
		{-1, 1, false},
	}
	for _, tt := range tests {
		if got := IsOverdue(tt.balance, tt.daysPastDue); got != tt.want {
			t.Errorf("IsOverdue(%d, %d) = %t, want %t", tt.balance, tt.daysPastDue, got, tt.want)
		}
	}
}

func TestLateFeeCents(t *testing.T) {
	tests := []struct {
		balance     int64
		daysPastDue int
		want        int64
	}{
		{100, 31, 500},
		{100, 29, 0},
		{-1, 31, 0},
	}
	for _, tt := range tests {
		if got := LateFeeCents(tt.balance, tt.daysPastDue); got != tt.want {
			t.Errorf("LateFeeCents(%d, %d) = %d, want %d", tt.balance, tt.daysPastDue, got, tt.want)
		}
	}
}
