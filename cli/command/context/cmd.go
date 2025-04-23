package context

import (
	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/spf13/cobra"
)

// NewContextCommand returns the context cli subcommand
func NewContextCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage contexts",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
	}
	cmd.AddCommand(
		newCreateCommand(dockerCli),
		newListCommand(dockerCli),
		newUseCommand(dockerCli),
		newExportCommand(dockerCli),
		newImportCommand(dockerCli),
		newRemoveCommand(dockerCli),
		newUpdateCommand(dockerCli),
		newInspectCommand(dockerCli),
		newShowCommand(dockerCli),
	)
	return cmd
}
