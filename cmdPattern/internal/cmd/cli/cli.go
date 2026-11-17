package cli

import (
	"context"
	"log/slog"
	"time"

	"cmdPattern/internal/cmd/cli/command"
	"cmdPattern/internal/cmd/cli/consumer"

	"github.com/spf13/cobra"
)

// Setup registers command groups with the root command.
func Setup(ctx context.Context) *cobra.Command {
	rootCmd := newRootCmd(ctx)

	rootCmd.AddCommand(
		command.New().Register(),
		consumer.New().Register(),
	)

	return rootCmd
}

// newRootCmd creates the root command with base configuration.
func newRootCmd(ctx context.Context) *cobra.Command {
	var (
		verbose bool
		start   time.Time
		rootCmd = &cobra.Command{
			Use:           "cli",
			Short:         "Cli",
			Version:       command.Version(),
			SilenceErrors: true,
			SilenceUsage:  true,
			PersistentPreRun: func(cmd *cobra.Command, _ []string) {
				if verbose {
					slog.InfoContext(cmd.Context(), "started command", "command", cmd.CommandPath())
				}
				start = time.Now()
			},
			PersistentPostRun: func(cmd *cobra.Command, _ []string) {
				if verbose {
					slog.InfoContext(cmd.Context(), "ended command", "executionTime", time.Since(start).Seconds())
				}
			},
		}
	)

	rootCmd.SetContext(ctx)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "increase logs verbose to debug")

	return rootCmd
}
