package patch

import (
	"fmt"

	"github.com/spf13/cobra"

	patchops "cmdPattern/internal/patch"
)

func removeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove a patch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), patchops.Remove())
			return err
		},
	}
}
