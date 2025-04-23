package system

import (
	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/spf13/cobra"
)

// NewSystemCommand returns a cobra command for `system` subcommands
func NewSystemCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Manage Docker",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
	}
	cmd.AddCommand(
		NewEventsCommand(dockerCli),
		NewInfoCommand(dockerCli),
		newDiskUsageCommand(dockerCli),
		newPruneCommand(dockerCli),
		newDialStdioCommand(dockerCli),
	)

	return cmd
}
