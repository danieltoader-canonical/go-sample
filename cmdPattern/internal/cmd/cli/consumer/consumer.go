package consumer

import (
	"github.com/spf13/cobra"
)

// Consumer is a struct that holds the command registration logic.
// It is used to group related commands under a common namespace, allowing for better organization
// and easier access to functionality within the CLI.
type Consumer struct{}

// New creates a new instance of Consumer.
// This function initializes a Consumer struct, which is used to register subcommands
// for a CLI application, allowing for modular command
// registration and management within the CLI framework.
func New() *Consumer {
	return &Consumer{}
}

// Register registers the command and its subcommands.
// It returns a *cobra.Command that can be used in the CLI application.
//
// The commands are registered under the "consumer" group, which can be invoked
// using the CLI as `cli consumer <subcommand>`.
func (cons *Consumer) Register() *cobra.Command {
	consumers := &cobra.Command{
		Use:   "consumer",
		Short: "Consumers",
	}
	consumers.AddCommand(sampleConsumer())

	return consumers
}
