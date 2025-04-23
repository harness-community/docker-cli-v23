package checkpoint

import (
	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/spf13/cobra"
)

// NewCheckpointCommand returns the `checkpoint` subcommand (only in experimental)
func NewCheckpointCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage checkpoints",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
		Annotations: map[string]string{
			"experimental": "",
			"ostype":       "linux",
			"version":      "1.25",
		},
	}
	cmd.AddCommand(
		newCreateCommand(dockerCli),
		newListCommand(dockerCli),
		newRemoveCommand(dockerCli),
	)
	return cmd
}
