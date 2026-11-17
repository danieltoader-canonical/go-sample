package sys

import (
	"context"
	"log/slog"
	"os"
)

// CheckErrorWithExit checks if the error is not nil and logs it, then exits the application with status code 1.
func CheckErrorWithExit(ctx context.Context, err error) {
	if err != nil {
		slog.ErrorContext(
			ctx,
			"exited with error",
			"fatal", true,
			"error", err,
		)
		os.Exit(1)
	}
}
