// Package invoice contains small invoice rules for mutation testing.
package invoice

// IsOverdue reports whether an unpaid balance is past its due date.
// Zero days means the invoice is due today; negative days mean it is not due yet.
func IsOverdue(balanceCents int64, daysPastDue int) bool {
	return balanceCents > 0 && daysPastDue > 0
}

// LateFeeCents returns a fixed fee for an unpaid invoice over 30 days late.
func LateFeeCents(balanceCents int64, daysPastDue int) int64 {
	if balanceCents > 0 && daysPastDue > 30 {
		return 500
	}
	return 0
}
