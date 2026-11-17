package command

import (
	"github.com/spf13/cobra"

	"cmdPattern/internal/cmd/cli/command/patch"
)

// Command is a struct that holds the command registration logic.
// It is used to group related commands under a common namespace, allowing for better organization
// and easier access to functionality within the CLI.
type Command struct{}

// New creates a new instance of Command.
// This function initializes a Command struct, which is used to register subcommands
// for a CLI application, allowing for modular command
// registration and management within the CLI framework.
func New() *Command { return &Command{} }

// Register registers the command and its subcommands.
// It returns a *cobra.Command that can be used in the CLI application.
//
// The commands are registered under the "command" group, which can be invoked
// using the CLI as `cli command <subcommand>`.
func (cmd *Command) Register() *cobra.Command {
	commands := &cobra.Command{
		Use:   "command",
		Short: "Commands",
	}

	commands.AddCommand(
		getVersionCommand(),
		patch.New().Register(),
	)

	return commands
}
