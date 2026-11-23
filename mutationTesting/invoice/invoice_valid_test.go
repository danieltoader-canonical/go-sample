//go:build !weak

package invoice

import "testing"

func TestIsOverdue(t *testing.T) {
	tests := []struct {
		name        string
		balance     int64
		daysPastDue int
		want        bool
	}{
		{"overdue", 100, 1, true},
		{"due today", 100, 0, false},
		{"not due yet", 100, -1, false},
		{"paid", 0, 1, false},
		{"credit balance", -1, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOverdue(tt.balance, tt.daysPastDue); got != tt.want {
				t.Errorf("IsOverdue(%d, %d) = %t, want %t", tt.balance, tt.daysPastDue, got, tt.want)
			}
		})
	}
}

func TestLateFeeCents(t *testing.T) {
	tests := []struct {
		name        string
		balance     int64
		daysPastDue int
		want        int64
	}{
		{"31 days late", 100, 31, 500},
		{"30 days late", 100, 30, 0},
		{"29 days late", 100, 29, 0},
		{"paid", 0, 31, 0},
		{"credit balance", -1, 31, 0},
		{"due today", 100, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LateFeeCents(tt.balance, tt.daysPastDue); got != tt.want {
				t.Errorf("LateFeeCents(%d, %d) = %d, want %d", tt.balance, tt.daysPastDue, got, tt.want)
			}
		})
	}
}
