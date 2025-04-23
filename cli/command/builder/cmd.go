package builder

import (
	"github.com/spf13/cobra"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/image"
)

// NewBuilderCommand returns a cobra command for `builder` subcommands
func NewBuilderCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "builder",
		Short:       "Manage builds",
		Args:        cli.NoArgs,
		RunE:        command.ShowHelp(dockerCli.Err()),
		Annotations: map[string]string{"version": "1.31"},
	}
	cmd.AddCommand(
		NewPruneCommand(dockerCli),
		image.NewBuildCommand(dockerCli),
	)
	return cmd
}
