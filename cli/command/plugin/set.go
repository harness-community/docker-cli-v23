package plugin

import (
	"context"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/spf13/cobra"
)

func newSetCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set PLUGIN KEY=VALUE [KEY=VALUE...]",
		Short: "Change settings for a plugin",
		Args:  cli.RequiresMinArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return dockerCli.Client().PluginSet(context.Background(), args[0], args[1:])
		},
	}

	return cmd
}
