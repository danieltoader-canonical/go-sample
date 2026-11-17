package patch

import (
	"fmt"

	"github.com/spf13/cobra"

	patchops "cmdPattern/internal/patch"
)

func promoteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "promote",
		Short: "Promote a patch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), patchops.Promote())
			return err
		},
	}
}
