package consumer

import (
	"fmt"

	"github.com/spf13/cobra"
)

func sampleConsumer() *cobra.Command {
	return &cobra.Command{
		Use:   "sample",
		Short: "Run the sample consumer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Sample consumer")
			return err
		},
	}
}
