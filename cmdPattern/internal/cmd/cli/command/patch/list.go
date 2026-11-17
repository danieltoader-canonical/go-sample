package patch

import (
	"fmt"

	"github.com/spf13/cobra"

	patchops "cmdPattern/internal/patch"
)

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List patches",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), patchops.List())
			return err
		},
	}
}
