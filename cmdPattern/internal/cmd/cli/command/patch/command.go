package patch

import (
	"github.com/spf13/cobra"
)

// Command registers patch-related CLI commands.
type Command struct{}

// New creates a patch command registrar.
func New() *Command { return &Command{} }

// Register returns the patch command group, available as
// `cli command patch <subcommand>`.
func (cmd *Command) Register() *cobra.Command {
	commands := &cobra.Command{
		Use:   "patch",
		Short: "Patch operations",
	}

	commands.AddCommand(
		listCommand(),
		infoCommand(),
		promoteCommand(),
		removeCommand(),
	)

	return commands
}
