package plugin

import (
	"context"

	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/inspect"
	flagsHelper "github.com/DevanshMathur19/docker-cli-v23/cli/flags"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	pluginNames []string
	format      string
}

func newInspectCommand(dockerCli command.Cli) *cobra.Command {
	var opts inspectOptions

	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] PLUGIN [PLUGIN...]",
		Short: "Display detailed information on one or more plugins",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.pluginNames = args
			return runInspect(dockerCli, opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.format, "format", "f", "", flagsHelper.InspectFormatHelp)
	return cmd
}

func runInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()
	getRef := func(ref string) (interface{}, []byte, error) {
		return client.PluginInspectWithRaw(ctx, ref)
	}

	return inspect.Inspect(dockerCli.Out(), opts.pluginNames, opts.format, getRef)
}
