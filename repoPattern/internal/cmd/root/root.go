package root

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"repoPattern/internal/dbstore/customer"

	"github.com/spf13/cobra"
)

func NewRootCommand(newRepository func(context.Context) (customer.Repository, func(), error)) *cobra.Command {
	var customerID int
	var name string
	cmd := &cobra.Command{
		Use:           "repoPattern",
		Short:         "Read a customer or update their name",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if customerID <= 0 {
				return fmt.Errorf("customer-id must be positive")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()
			repo, cleanup, err := newRepository(ctx)
			if err != nil {
				return err
			}
			defer cleanup()

			if cmd.Flags().Changed("name") {
				if _, err := repo.UpdateCustomerName(ctx, customerID, name); err != nil {
					return err
				}
			}
			record := repo.GetCustomer(ctx, customerID)
			if record.CustomerId == 0 {
				return fmt.Errorf("customer %d could not be loaded; see logs for database errors", customerID)
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(record)
		},
	}
	cmd.Flags().IntVar(&customerID, "customer-id", 1, "customer ID to read or update")
	cmd.Flags().StringVar(&name, "name", "", "update the customer name before reading")
	return cmd
}
