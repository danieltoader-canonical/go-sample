package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"cmdPattern/pkg/version"
)

func Version() string {
	return version.GetBuildInfo().Version
}

func getVersionCommand() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "get-version",
		Short: "Get binary version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			bi := version.GetBuildInfo()
			v := bi.String()
			if short {
				v = bi.Short(cmd.Context())
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), v)

			return err
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "print the version number")

	return cmd
}
