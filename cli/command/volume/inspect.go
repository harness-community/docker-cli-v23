package volume

import (
	"context"

	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/completion"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/inspect"
	flagsHelper "github.com/DevanshMathur19/docker-cli-v23/cli/flags"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	format string
	names  []string
}

func newInspectCommand(dockerCli command.Cli) *cobra.Command {
	var opts inspectOptions

	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] VOLUME [VOLUME...]",
		Short: "Display detailed information on one or more volumes",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.names = args
			return runInspect(dockerCli, opts)
		},
		ValidArgsFunction: completion.VolumeNames(dockerCli),
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "", flagsHelper.InspectFormatHelp)

	return cmd
}

func runInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()

	ctx := context.Background()

	getVolFunc := func(name string) (interface{}, []byte, error) {
		i, err := client.VolumeInspect(ctx, name)
		return i, nil, err
	}

	return inspect.Inspect(dockerCli.Out(), opts.names, opts.format, getVolFunc)
}
