package patch

import (
	"fmt"

	"github.com/spf13/cobra"

	patchops "cmdPattern/internal/patch"
)

func infoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show patch information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), patchops.Info())
			return err
		},
	}
}
